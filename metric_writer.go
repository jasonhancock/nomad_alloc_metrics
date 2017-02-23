package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type MetricWriter interface {
	WriteMetric(metric string, value int, ts time.Time) error
}

type GraphiteMetricWriter struct {
	conn net.Conn
}

func NewGraphiteMetricWriter(server string) (*GraphiteMetricWriter, error) {

	d := &net.Dialer{Timeout: 5 * time.Second}
	c, err := d.Dial("tcp", server)
	if err != nil {
		return nil, err
	}

	return &GraphiteMetricWriter{conn: c}, nil
}

func (m *GraphiteMetricWriter) Close() {
	m.conn.Close()
}

func (m *GraphiteMetricWriter) WriteMetric(metric string, value int, ts time.Time) error {
	str := fmt.Sprintf("%s %d %d\n", metric, value, ts.Unix())
	log.Print(str)
	_, err := m.conn.Write([]byte(str))
	return err
}

type LoggingMetricWriter struct{}

func (m *LoggingMetricWriter) WriteMetric(metric string, value int, ts time.Time) error {
	str := fmt.Sprintf("%s %d %d\n", metric, value, ts.Unix())
	log.Print(str)
	return nil
}
