package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"os"
	"syscall"
	"encoding/json"
)

type Metricser interface {
	Execute(address string) error
	RecordMessage(msgType string)
	GetMetrics() MetricsData
	ResolveSignals(c chan os.Signal)
}

type MetricsData struct {
	MsgAck   int `json:"msg_ack"`
	MsgNak   int `json:"msg_nak"`
	MsgReq   int `json:"msg_req"`
	MsgTotal int `json:"msg_total"`
}

var MetricsInstance Metricser = &Metrics{}

type Metrics struct{}

var msg = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "msg_counter",
		Help: "Total number of messages",
	},
	[]string{"type"},
)

func init() {
	prometheus.MustRegister(msg)
}

func (m *Metrics) Execute(address string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(address, nil)
}

func (m *Metrics) RecordMessage(msgType string) {
	msg.WithLabelValues(msgType).Inc()
}

// Switch to map
func (m *Metrics) GetMetrics() MetricsData {
	metricsFamilies, _ := prometheus.DefaultGatherer.Gather()
	data := MetricsData{}
	for _, metricFamily := range metricsFamilies {
		if metricFamily.GetName() == "msg_counter" {
			for _, metric := range metricFamily.Metric {
				value := int(metric.GetCounter().GetValue())
				for _, label := range metric.GetLabel() {
					switch label.GetValue() {
					case "ACK":
						data.MsgAck = value
					case "REQ":
						data.MsgReq = value
					case "NAK":
						data.MsgNak = value
					}
				}
				data.MsgTotal += value
			}
		}
	}
	return data
}

func (m *Metrics) ResolveSignals(c chan os.Signal) {
	for {
		sig := <-c
		switch sig {
		case syscall.SIGUSR1:
			json, _ := json.Marshal(m.GetMetrics())
			println(string(json))
		default:
			println("Not supported")
		}
	}
}
