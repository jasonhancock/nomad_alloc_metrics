package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"github.com/pkg/errors"
)

func main() {

	var (
		graphite    = flag.String("graphite-addr", "127.0.0.1:2003", "host and port of carbon server")
		nomadAddr   = flag.String("addr", "http://127.0.0.1:4646", "The address of the Nomad server")
		tlsCert     = flag.String("tls-cert", "", "TLS certificate to use when connecting to Nomad")
		tlsKey      = flag.String("tls-key", "", "TLS key to use when connecting to Nomad")
		tlsCaCert   = flag.String("tls-ca-cert", "", "TLS CA cert to use to validate the Nomad server certificate")
		tlsInsecure = flag.Bool("tls-insecure", false, "Whether or not to validate the server certificate")
	)
	flag.Parse()

	cfg := api.DefaultConfig()
	cfg.Address = *nomadAddr
	if *tlsCert != "" && *tlsKey != "" {
		cfg.TLSConfig = &api.TLSConfig{
			CACert:     *tlsCaCert,
			ClientCert: *tlsCert,
			ClientKey:  *tlsKey,
			Insecure:   *tlsInsecure,
		}
	}
	c, err := api.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	nodeId, region, datacenter, err := getNodeInfo(c)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("node_id=%s region=%s datacenter=%s", nodeId, region, datacenter)

	total, used, err := getResources(c, nodeId)
	if err != nil {
		log.Fatal(err)
	}
	ts := time.Now()

	mw, err := NewGraphiteMetricWriter(*graphite)
	if err != nil {
		log.Fatal(err)
	}
	defer mw.Close()

	prefix := "nomad"
	host, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	hostname := strings.Replace(host, ".", "_", -1)
	pfx := fmt.Sprintf("%s.%s.%s.%s.", prefix, region, datacenter, hostname)

	writeMetrics(pfx, "total", total, mw, ts)
	writeMetrics(pfx, "allocated", used, mw, ts)
}

func writeMetrics(prefix, rType string, r *api.Resources, mw MetricWriter, ts time.Time) {
	mw.WriteMetric(prefix+"CPU_"+rType, r.CPU, ts)
	mw.WriteMetric(prefix+"MemoryMB_"+rType, r.MemoryMB, ts)
	mw.WriteMetric(prefix+"DiskMB_"+rType, r.DiskMB, ts)
	mw.WriteMetric(prefix+"IOPS_"+rType, r.IOPS, ts)
}

func getNodeInfo(client *api.Client) (string, string, string, error) {
	info, err := client.Agent().Self()
	if err != nil {
		return "", "", "", errors.Wrap(err, "querying agent info")
	}
	var stats map[string]interface{}
	stats, _ = info["stats"]
	clientStats, ok := stats["client"].(map[string]interface{})
	if !ok {
		return "", "", "", errors.New("Nomad not running in client mode")
	}

	nodeID, ok := clientStats["node_id"].(string)
	if !ok {
		return "", "", "", errors.New("Failed to determine node ID")
	}

	var config map[string]interface{}

	config, _ = info["config"]
	region := config["Region"].(string)
	dc := config["Datacenter"].(string)

	return nodeID, region, dc, nil
}

func getResources(client *api.Client, id string) (*api.Resources, *api.Resources, error) {
	node, _, err := client.Nodes().Info(id, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "querying node")
	}

	// Total available resources
	total := &api.Resources{}

	r := node.Resources
	res := node.Reserved
	if res == nil {
		res = &api.Resources{}
	}
	total.CPU = r.CPU - res.CPU
	total.MemoryMB = r.MemoryMB - res.MemoryMB
	total.DiskMB = r.DiskMB - res.DiskMB
	total.IOPS = r.IOPS - res.IOPS

	// Get allocated resources
	allocated := &api.Resources{}
	nodeAllocs, _, err := client.Nodes().Allocations(id, nil)
	if err != nil {
		return nil, nil, err
	}

	// Filter list to only running allocations
	for _, alloc := range nodeAllocs {
		if alloc.ClientStatus != "running" {
			continue
		}
		allocated.CPU += alloc.Resources.CPU
		allocated.MemoryMB += alloc.Resources.MemoryMB
		allocated.DiskMB += alloc.Resources.DiskMB
		allocated.IOPS += alloc.Resources.IOPS
	}

	return total, allocated, nil
}
