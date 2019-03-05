package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	Failed = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "failed",
		Namespace: "expressionist",
		Help:      "number of resources failed validation",
	})
	Validations = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "validations",
		Namespace: "expressionist",
		Help:      "number of resources validated",
	})
)

func init() {
	prometheus.MustRegister(Failed)
	prometheus.MustRegister(Validations)
}

// Serve health and metric requests forever.
func Serve(addr, metrics, ready, alive string) {
	h := http.NewServeMux()
	h.Handle(metrics, promhttp.Handler())
	log.Infof("Metrics and status server started on %s", addr)
	log.Infof("Serving metrics on %s", metrics)
	log.Info(http.ListenAndServe(addr, h))
}
