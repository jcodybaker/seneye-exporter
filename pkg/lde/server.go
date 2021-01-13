package lde

import (
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// Server is an HTTP server which implements the Seneye LDE protocol.
type Server struct {
	lastLDEs map[string]*LDE
	lock     sync.Mutex
	secrets  map[string][]byte

	log *logrus.Entry
}

// ServeHTTP implements an http.Handler for the LDE server.
func (l *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ll := l.log.WithFields(logrus.Fields{
		"http_method": r.Method,
		"request_uri": r.RequestURI,
		"remote_addr": r.RemoteAddr,
	})
	ll.Debug("HTTP LDE request received")
	for k, values := range r.Header {
		for _, v := range values {
			ll.WithFields(logrus.Fields{"http_header_name": k, "http_header_value": v}).Trace("processing HTTP LDE request header")
		}
	}
	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		ll.WithError(err).Error("reading LDE body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	lde, err := FromRequestBody(msg, l.secrets)
	if err != nil {
		ll.WithError(err).Error("parsing LDE body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	l.lock.Lock()
	l.lastLDEs[lde.SUD.ID] = lde
	l.lock.Unlock()
	ll.WithFields(logrus.Fields{
		"lde_version":            lde.Version,
		"sud_id":                 lde.SUD.ID,
		"sud_name":               lde.SUD.Name,
		"sud_type":               lde.SUD.Type.String(),
		"sud_data_temperature":   lde.SUD.Data.Temperature,
		"sud_data_ph":            lde.SUD.Data.PH,
		"sud_data_nh3":           lde.SUD.Data.NH3,
		"sud_data_kelvin":        lde.SUD.Data.Kelvin,
		"sud_data_lux":           lde.SUD.Data.Lux,
		"sud_data_par":           lde.SUD.Data.PAR,
		"sud_status_water":       lde.SUD.Data.Status.Water,
		"sud_status_temperature": lde.SUD.Data.Status.Temperature,
		"sud_status_ph":          lde.SUD.Data.Status.PH,
		"sud_status_nh3":         lde.SUD.Data.Status.NH3,
		"sud_status_slide":       lde.SUD.Data.Status.Slide,
		"sud_status_kelvin":      lde.SUD.Data.Status.Kelvin,
	}).Debug("LDE event received")
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

// WithLog sets the logger.
func WithLog(log *logrus.Entry) ServerOption {
	return func(s *Server) {
		s.log = log
	}
}

// WithPrometheus registers the server with a prometheus registry
func WithPrometheus(reg prometheus.Registerer) ServerOption {
	return func(s *Server) {
		reg.MustRegister(s)
	}
}
