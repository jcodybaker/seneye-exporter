package lde

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/hlog"
)

// Server is an HTTP server which implements the Seneye LDE protocol.
type Server struct {
	lastLDEs map[string]*LDE
	lock     sync.Mutex
	secrets  map[string][]byte
}

// ServeHTTP implements an http.Handler for the LDE server.
func (l *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ll := hlog.FromRequest(r)
	for k, values := range r.Header {
		for _, v := range values {
			ll.Trace().Str("http_header_name", k).Str("http_header_value", v).Msg("processing HTTP LDE request header")
		}
	}
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ll.Error().Err(err).Msg("reading LDE body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	lde, err := FromRequestBody(msg, l.secrets)
	if err != nil {
		ll.Error().Err(err).Msg("parsing LDE body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	l.lock.Lock()
	l.lastLDEs[lde.SUD.ID] = lde
	l.lock.Unlock()
	ll.Debug().
		Str("lde_version", lde.Version).
		Str("sud_id", lde.SUD.ID).
		Str("sud_name", lde.SUD.Name).
		Str("sud_type", lde.SUD.Type.String()).
		Float64("sud_data_temperature", lde.SUD.Data.Temperature).
		Float64("sud_data_ph", lde.SUD.Data.PH).
		Float64("sud_data_nh3", lde.SUD.Data.NH3).
		Int("sud_data_kelvin", lde.SUD.Data.Kelvin).
		Int("sud_data_lux", lde.SUD.Data.Lux).
		Int("sud_data_par", lde.SUD.Data.PAR).
		Int("sud_status_water", lde.SUD.Data.Status.Water).
		Int("sud_status_temperature", lde.SUD.Data.Status.Temperature).
		Int("sud_status_ph", lde.SUD.Data.Status.PH).
		Int("sud_status_nh3", lde.SUD.Data.Status.NH3).
		Int("sud_status_slide", lde.SUD.Data.Status.Slide).
		Int("sud_status_kelvin", lde.SUD.Data.Status.Kelvin).
		Msg("LDE event received")
	w.WriteHeader(http.StatusNoContent)
}

// ServerOption describes a func which implements the functional option pattern for the LDE Server.
type ServerOption func(*Server)

// NewServer creates a new LDE Server with the provided options.
func NewServer(options ...ServerOption) *Server {
	s := &Server{
		lastLDEs: make(map[string]*LDE),
	}
	for _, o := range options {
		o(s)
	}
	return s
}

// WithServer sets the JWT validation secrets used to verify the authenticity of an LDE request.
// Secrets is a map of SUD ID to JWT signing secret, allowing one server to record LDE events for
// multiple SUDs / Seneye accounts. A default secret for all unspecified SUDs can be set w/ the
// empty-string for key.
func WithSecrets(secrets map[string][]byte) ServerOption {
	return func(s *Server) {
		s.secrets = secrets
	}
}

// WithPrometheus registers the server with a prometheus registry
func WithPrometheus(reg prometheus.Registerer) ServerOption {
	return func(s *Server) {
		reg.MustRegister(s)
	}
}
