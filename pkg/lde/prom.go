package lde

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	labels = []string{"id", "name", "sud_type"}

	tempDesc = prometheus.NewDesc(
		"temperature_celsius",
		"Water temperature in celsius",
		labels, nil,
	)
	phDesc = prometheus.NewDesc(
		"ph",
		"Water pH",
		labels, nil,
	)
	ammoniaDesc = prometheus.NewDesc(
		"ammonia",
		"PPM Water NH3 free ammonia",
		labels, nil,
	)
	kelvinDesc = prometheus.NewDesc(
		"light_kelvin",
		"Kelvin is the numeric Correlated Color Temperature value of the colour temperature in degrees Kelvin.",
		labels, nil,
	)
	luxDesc = prometheus.NewDesc(
		"light_lux",
		"Lux describes the intensity of the light observed in the tank. ",
		labels, nil,
	)
	parDesc = prometheus.NewDesc(
		"light_par",
		"PAR describes the photosynthetic active radiation is a measurement of light power between 400nm and 700nm.",
		labels, nil,
	)

	statusWaterDesc = prometheus.NewDesc(
		"seneye_status_water",
		"Water is 1 if the SUD is submerged in water, 0 otherwise.",
		labels, nil,
	)
	statusTemperatureDesc = prometheus.NewDesc(
		"seneye_status_temperature",
		"Temperature is 0 if the temperature is within limits, 1 otherwise.",
		labels, nil,
	)
	statusPhDesc = prometheus.NewDesc(
		"seneye_status_ph",
		"PH is 0 if the pH is within limits, 1 otherwise.",
		labels, nil,
	)
	statusAmmoniaDesc = prometheus.NewDesc(
		"seneye_status_ammonia",
		"Ammonia (NH3) is 0 if the free ammonia is within limits, 1 otherwise.",
		labels, nil,
	)
	statusSlideDesc = prometheus.NewDesc(
		"seneye_status_slide",
		"Slide is 0 if the slide is correctly installed and unexpired, 1 otherwise.",
		labels, nil,
	)
	statusKelvinDesc = prometheus.NewDesc(
		"seneye_status_kelvin",
		"Kelvin is 0 if the Kelvin measurement is within limits, 1 otherwise.",
		labels, nil,
	)
)

// Describe implements prometheus.Collector.
func (l *Server) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(l, ch)
}

// Collect implements prometheus.Collector.
func (l *Server) Collect(ch chan<- prometheus.Metric) {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, l := range l.lastLDEs {
		labels := []string{l.SUD.ID, l.SUD.Name, l.SUD.Type.String()}
		t := time.Unix(l.SUD.Timestamp, 0)
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			tempDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Temperature),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			phDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.PH),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			ammoniaDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.NH3),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			kelvinDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Kelvin),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			luxDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Lux),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			parDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.PAR),
			labels...,
		))

		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusWaterDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.Water),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusTemperatureDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.Temperature),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusPhDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.PH),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusAmmoniaDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.NH3),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusSlideDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.Slide),
			labels...,
		))
		ch <- prometheus.NewMetricWithTimestamp(t, prometheus.MustNewConstMetric(
			statusKelvinDesc,
			prometheus.GaugeValue,
			float64(l.SUD.Data.Status.Kelvin),
			labels...,
		))
	}
}
