package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/jcodybaker/seneye-exporter/pkg/lde"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ctx     context.Context
	cfgFile string

	rootCmd = &cobra.Command{
		Use:    "seneye-exporter",
		Short:  "Listen to seneye LDE events and export them to metrics storage.",
		PreRun: configureLog,
		Run:    rootExecute,
	}
)

func init() {
	ctx = context.Background()
	cobra.OnInitialize(initConfig)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().String("log-format", "text", `log format: "json", "text"`)
	viper.BindPFlag("log-format", rootCmd.PersistentFlags().Lookup("log-format"))
	viper.SetDefault("log-format", "text")
	rootCmd.PersistentFlags().String("log-level", zerolog.DebugLevel.String(),
		fmt.Sprintf("log level: %q %q %q \n%q %q %q %q",
			zerolog.TraceLevel.String(),
			zerolog.DebugLevel.String(),
			zerolog.InfoLevel.String(),
			zerolog.WarnLevel.String(),
			zerolog.ErrorLevel.String(),
			zerolog.FatalLevel.String(),
			zerolog.PanicLevel.String(),
		))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("log-level", zerolog.DebugLevel.String())

	rootCmd.Flags().Uint16("prom-port", 9090, "Port for prometheus metrics server")
	viper.BindPFlag("prom-port", rootCmd.Flags().Lookup("prom-port"))
	viper.SetDefault("prom-port", uint16(9090))

	rootCmd.Flags().Uint16("lde-port", 8080, "Port for LDE server")
	viper.BindPFlag("lde-port", rootCmd.Flags().Lookup("lde-port"))
	viper.SetDefault("lde-port", uint16(8080))

	rootCmd.Flags().StringSlice("lde-secret", nil, `Secret used to validate LDE message authenticity. --lde-secret may be specified
multiple times if paired with the SUD ID. (ex. --lde-secret=DEFAULT_SECRET, or
--lde-secret=EXAMPLE_SUD_ID=SECRET1 --lde-secret=OTHER_SUD_ID=SECRET2)`)
	viper.BindPFlag("lde-secret", rootCmd.Flags().Lookup("lde-secret"))
}

func main() {
	rootCmd.Execute()
}

func rootExecute(cmd *cobra.Command, args []string) {
	secrets := parseSecrets(cmd)

	promPort := viper.GetUint("prom-port")
	if promPort > 0xFFFF {
		log.Fatal().Str("prom_port", viper.GetString("prom-port")).Msg("invalid prom-port")
	}
	ldePort := viper.GetUint("lde-port")
	if ldePort > 0xFFFF {
		log.Fatal().Str("lde_port", viper.GetString("lde-port")).Msg("invalid lde-port")
	}

	promRegistry := prometheus.NewPedanticRegistry()
	promRegistry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	ldeServer := lde.NewServer(
		lde.WithPrometheus(promRegistry),
		lde.WithSecrets(secrets),
	)

	ldeMux := http.NewServeMux()
	promMux := ldeMux
	if promPort != ldePort {
		promMux = http.NewServeMux()
	}

	ldeMux.Handle("/lde", ldeServer)
	promMux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))

	eg, runCtx := errgroup.WithContext(ctx)
	var ldeHTTP, promHTTP *http.Server
	eg.Go(func() error {
		l := log.With().Uint("server_port", ldePort).Logger()
		ldeHTTP = &http.Server{
			Addr:    fmt.Sprintf(":%d", ldePort),
			Handler: logHandler(ldeMux),
			BaseContext: func(net.Listener) context.Context {
				return l.WithContext(ctx)
			},
		}
		l.Info().Msg("starting lde http server")
		err := ldeHTTP.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Uint("lde-port", ldePort).Msg("failed to start LDE http server")
			return err
		}
		return nil
	})

	if promPort != ldePort {
		eg.Go(func() error {
			l := log.With().Uint("server_port", ldePort).Logger()
			promHTTP = &http.Server{
				Addr:    fmt.Sprintf(":%d", promPort),
				Handler: logHandler(promMux),
				BaseContext: func(net.Listener) context.Context {
					return l.WithContext(ctx)
				},
			}
			l.Info().Msg("starting prometheus http server")
			err := promHTTP.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Error().Err(err).Uint("prom-port", promPort).Msg("failed to start prometheus http server")
				return err
			}
			return nil
		})
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		log.Info().Msg("Got SIGINT; shutting down")
	case <-runCtx.Done():
		// One of the listeners failed, but these log their own error.
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var shutdownEG errgroup.Group
	shutdownEG.Go(func() error {
		return ldeHTTP.Shutdown(ctx)
	})
	if promPort != ldePort {
		shutdownEG.Go(func() error {
			return promHTTP.Shutdown(ctx)
		})
	}
	eg.Wait() // We've now told all servers to shutdown, wait for their goroutines to exit.
	if err := shutdownEG.Wait(); err != nil {
		log.Warn().Err(err).Msg("shutting down HTTP servers")
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile == "" {
		return
	}
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Fatal().Err(err).Str("config", viper.ConfigFileUsed()).Msg("failed to read config file")
	}
}

func configureLog(cmd *cobra.Command, args []string) {
	logFormat := viper.GetString("log-format")
	switch logFormat {
	case "text":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	case "json": // zerolog default, but we default to text
	default:
		log.Fatal().Str("log_format", logFormat).Msg("unknown log format")
	}
	levelS := viper.GetString("log-level")
	level, err := zerolog.ParseLevel(levelS)
	if err != nil {
		log.Fatal().Str("log_level", levelS).Msg("unknown log level")
	}
	zerolog.SetGlobalLevel(level)
	ctx = log.Logger.WithContext(ctx)
}

func parseSecrets(cmd *cobra.Command) map[string][]byte {
	out := make(map[string][]byte)
	secrets := viper.GetStringSlice("lde-secret")
	if len(secrets) == 0 {
		cmd.Usage()
		log.Fatal().Msg("lde-secret is required")
	}
	for _, s := range secrets {
		if s == "" {
			cmd.Usage()
			log.Fatal().Msg("invalid lde-secret")
		}
		splitSecret := strings.SplitN(s, "=", 1)
		if len(splitSecret) == 1 {
			if _, ok := out[""]; ok {
				log.Fatal().Msg("only one default secret may be provided")
			}
			out[""] = []byte(s)
			continue
		}
		if _, ok := out[splitSecret[0]]; ok {
			log.Fatal().Str("sud_id", splitSecret[0]).Msg("SUD ID had >1 secrets")
		}
		out[splitSecret[0]] = []byte(splitSecret[1])
	}
	return out
}

func logHandler(h http.Handler) http.Handler {
	h = hlog.NewHandler(log.Logger)(h)
	h = hlog.RemoteAddrHandler("ip")(h)
	h = hlog.UserAgentHandler("user_agent")(h)
	h = hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		l := log.Ctx(r.Context()).With().
			Int("http_status", status).
			Int("response_bytes", size).
			Dur("duration", duration).
			Str("http_method", r.Method).
			Str("request_uri", r.RequestURI).Logger()
		if status < 400 {
			l.Debug().Msg("request completed")
		} else {
			l.Warn().Msg("error processing request")
		}

	})(h)

	return h
}
