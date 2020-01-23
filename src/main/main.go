package main

import (
	"flag"
	kubernetes "github.com/danielr1996/prometheus-sd-kubernetes-node-metrics/kube-api"
	"github.com/danielr1996/prometheus-sd-kubernetes-node-metrics/prometheus"
	v1 "k8s.io/api/core/v1"
)

func main() {
	var environment string
	flag.StringVar(&environment, "environment", "kubernetes", "Generate config for 'local' or 'kubernetes'")
	flag.Parse()

	var nodes = kubernetes.GetNodes(environment)
	targetsConfig := GenerateTargets(nodes)

	prometheus.WriteTargetsConfig("/var/prometheus/targets.json", targetsConfig)
}

func GenerateTargets(nodes []v1.Node) []prometheus.TargetsConfig{
	nodeList := []string{}
	for _, node := range nodes {
		nodeList = append(nodeList, "kubernetes.default"+"/api/v1/nodes/"+node.Name+"/proxy/metrics/cadvisor")
	}
	targetsConfig := []prometheus.TargetsConfig{prometheus.TargetsConfig{Targets: nodeList}}
	return targetsConfig
}
