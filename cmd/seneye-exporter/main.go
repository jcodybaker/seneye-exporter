package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jcodybaker/seneye-exporter/pkg/lde"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var (
	promPort uint16
	ldePort  uint16
)

func main() {
	ll := logrus.NewEntry(logrus.New())
	promRegistry := prometheus.NewPedanticRegistry()
	promRegistry.MustRegister(
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		prometheus.NewGoCollector(),
	)

	ldeServer := lde.NewServer(
		lde.WithLog(ll),
		lde.WithPrometheus(promRegistry),
	)

	ldeMux := http.NewServeMux()
	promMux := ldeMux
	if promPort != ldePort {
		promMux = http.NewServeMux()
	}

	ldeMux.Handle("/lde", ldeServer)
	promMux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))

	eg, ctx := errgroup.WithContext(context.Background())
	var ldeHTTP, promHTTP *http.Server
	eg.Go(func() error {
		ldeHTTP = &http.Server{
			Addr:    fmt.Sprintf(":%d", ldePort),
			Handler: ldeMux,
		}
		err := ldeHTTP.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			ll.WithError(err).Error("failed to start LDE http server")
			return err
		}
		return nil
	})

	if promPort != ldePort {
		eg.Go(func() error {
			promHTTP = &http.Server{
				Addr:    fmt.Sprintf(":%d", promPort),
				Handler: promMux,
			}
			err := promHTTP.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				ll.WithError(err).Error("failed to start prometheus http server")
				return err
			}
			return nil
		})
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	select {
	case <-sigint:
		ll.Info("Got SIGINT; shutting down")
	case <-ctx.Done():
		// One of the listeners failed, but these log their own error.
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	shutdownEG, ctx := errgroup.WithContext(ctx)
	shutdownEG.Go(func() error {
		return ldeHTTP.Shutdown(ctx)
	})
	if promPort != ldePort {
		shutdownEG.Go(func() error {
			return promHTTP.Shutdown(ctx)
		})
	}
	eg.Wait()
	if err := shutdownEG.Wait(); err != nil {
		ll.WithError(err).Warn("shutting down HTTP servers")
		os.Exit(1)
	}
}
