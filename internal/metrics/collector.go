package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/eliecharra/ghoma/internal/ghoma"
	"github.com/eliecharra/ghoma/protocol"
)

const ns = "ghoma"

type status struct {
	Switch                                     *float64
	Power, Energy, Voltage, Current, Frequency float64
	PowerMax, CosPhi                           float64

	LastContact map[string]time.Time
}

type Collector struct {
	Switch      *prometheus.Desc
	Power       *prometheus.Desc
	Energy      *prometheus.Desc
	Voltage     *prometheus.Desc
	Frequency   *prometheus.Desc
	Current     *prometheus.Desc
	PowerMax    *prometheus.Desc
	CosPhi      *prometheus.Desc
	LastContact *prometheus.Desc

	status map[string]*status
	mu     sync.RWMutex
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.Switch
	ch <- c.Power
	ch <- c.Energy
	ch <- c.Voltage
	ch <- c.Frequency
	ch <- c.Current
	ch <- c.PowerMax
	ch <- c.CosPhi
	ch <- c.LastContact
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for device, s := range c.status {
		labels := []string{device}
		if s.Switch != nil {
			ch <- prometheus.MustNewConstMetric(c.Switch, prometheus.GaugeValue, *s.Switch, labels...)
		}
		ch <- prometheus.MustNewConstMetric(c.Power, prometheus.GaugeValue, s.Power, labels...)
		ch <- prometheus.MustNewConstMetric(c.Energy, prometheus.CounterValue, s.Energy, labels...)
		ch <- prometheus.MustNewConstMetric(c.Voltage, prometheus.GaugeValue, s.Voltage, labels...)
		ch <- prometheus.MustNewConstMetric(c.Frequency, prometheus.GaugeValue, s.Frequency, labels...)
		ch <- prometheus.MustNewConstMetric(c.Current, prometheus.GaugeValue, s.Current, labels...)
		ch <- prometheus.MustNewConstMetric(c.PowerMax, prometheus.GaugeValue, s.PowerMax, labels...)
		ch <- prometheus.MustNewConstMetric(c.CosPhi, prometheus.GaugeValue, s.CosPhi, labels...)
		for k, v := range s.LastContact {
			labels := append(labels, k)
			ch <- prometheus.MustNewConstMetric(c.LastContact, prometheus.GaugeValue, time.Since(v).Seconds(), labels...)
		}
	}
}

func (c *Collector) HandleStatus(dev *ghoma.Device, msg protocol.Message) {
	var s *status
	if _, exist := c.status[dev.ID]; !exist {
		c.mu.Lock()
		c.status[dev.ID] = &status{
			LastContact: make(map[string]time.Time, 8),
		}
		c.mu.Unlock()
	}
	s = c.status[dev.ID]

	if msg.Status.Switch != nil {
		s.LastContact["switch"] = time.Now()
		var val float64
		if *msg.Status.Switch {
			val = 1
		}
		s.Switch = &val
	}
	if msg.Status.Energy != nil {
		val := float64(msg.Status.Energy.Value()) / 100
		s.LastContact[msg.Status.Energy.Kind()] = time.Now()
		switch msg.Status.Energy.Kind() {
		case "POWER":
			s.Power = val
		case "ENERGY":
			s.Energy = val
		case "VOLTAGE":
			s.Voltage = val
		case "FREQUENCY":
			s.Frequency = val
		case "CURRENT":
			s.Current = val
		case "MAX_POWER":
			s.PowerMax = val
		case "COSPHI":
			s.CosPhi = val
		}
	}
}

func NewCollector() *Collector {
	labels := []string{"device"}
	return &Collector{
		status: make(map[string]*status),
		mu:     sync.RWMutex{},
		Switch: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "switch", "state"),
			"reported switch status ( 0 = off, 1 = on)",
			labels,
			nil,
		),
		Power: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "power"),
			"", // TODO
			labels,
			nil,
		),
		Energy: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "energy"),
			"", // TODO
			labels,
			nil,
		),
		Voltage: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "voltage"),
			"", // TODO
			labels,
			nil,
		),
		Frequency: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "frequency"),
			"power line frequency (in hertz)",
			labels,
			nil,
		),
		Current: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "current"),
			"current in ampere",
			labels,
			nil,
		),
		PowerMax: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "power_max"),
			"", // TODO
			labels,
			nil,
		),
		CosPhi: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "energy", "cos_phi"),
			"", // TODO
			labels,
			nil,
		),

		LastContact: prometheus.NewDesc(
			prometheus.BuildFQName(ns, "last_contact", "seconds"),
			"last timestamp data were updated", // TODO
			append(labels, "metric_kind"),
			nil,
		),
	}
}
