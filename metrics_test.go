package main

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"net/http"
	"io/ioutil"
	"os"
)

type MetricsTestSuite struct {
	suite.Suite
}

func (s *MetricsTestSuite) SetupTest() {}

func TestMetricsUnitTestSuite(t *testing.T) {
	s := new(MetricsTestSuite)
	go MetricsInstance.Execute(":8003")
	suite.Run(t, s)
}

// RecordMessage

func (s *MetricsTestSuite) Test_GetMetrics_ReturnsData() {
	expectedAck := MetricsInstance.GetMetrics().MsgAck + 1
	expectedReq := MetricsInstance.GetMetrics().MsgReq + 2
	expectedNak := MetricsInstance.GetMetrics().MsgNak + 1

	MetricsInstance.RecordMessage("ACK")
	MetricsInstance.RecordMessage("NAK")
	MetricsInstance.RecordMessage("REQ")
	MetricsInstance.RecordMessage("REQ")

	actual := MetricsInstance.GetMetrics()

	s.Equal(expectedAck + expectedReq + expectedNak, actual.MsgTotal)
	s.Equal(expectedAck, actual.MsgAck)
	s.Equal(expectedNak, actual.MsgNak)
	s.Equal(expectedReq, actual.MsgReq)
}

// Execute

func (s *MetricsTestSuite) Test_Execute_RunsMetricsHandler() {
	MetricsInstance.RecordMessage("ACK")
	resp, err := http.Get("http://localhost:8003/metrics")
	defer resp.Body.Close()

	// /metrics was started started in TestMetricsUnitTestSuite()

	s.NoError(err)
	s.Equal(200, resp.StatusCode)
	body, _ := ioutil.ReadAll(resp.Body)
	actual := string(body)
	s.Contains(actual, `msg_counter{type="ACK"}`)
}

// Mocks

type MetricsMock struct {
	ExecuteMock       func(address string) error
	RecordMessageMock func(msgType string)
	GetMetricsMock    func() MetricsData
	ResolveSignalsMock func(c chan os.Signal)
}

func (m MetricsMock) Execute(address string) error {
	return m.ExecuteMock(address)
}

func (m MetricsMock) RecordMessage(msgType string) {
	m.RecordMessageMock(msgType)
}

func (m MetricsMock) GetMetrics() MetricsData {
	return m.GetMetricsMock()
}

func (m MetricsMock) ResolveSignals(c chan os.Signal) {
	m.ResolveSignals(c)
}
