package main

import (
	"net/http"

	"github.com/jcodybaker/seneye-exporter/pkg/lde"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
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

	mux := http.NewServeMux()
	mux.Handle("/lde", ldeServer)
	mux.Handle("/metrics", promhttp.HandlerFor(promRegistry, promhttp.HandlerOpts{}))

	s := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := s.ListenAndServe(); err != nil {
		ll.WithError(err).Fatal("starting HTTP server")
	}
}
