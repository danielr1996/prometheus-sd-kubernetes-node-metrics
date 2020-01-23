# prometheus-sd-kubernetes-node-metrics
Autodiscover kubernetes node metrics for prometheus

# Why?
Kubernetes Exposes Pod Metrics like cpu usage per pod and memory usage per pod by proxying the
cAdvisor metrics from ech node to `<kubernetes-api>/api/v1/nodes/<node-name>/proxy/metrics/cadvisor`
so there is no single endpoint where the pod level metrics can be scraped. So this project can be used
to generate a `targets.json` for prometheus to dynamicaly discover the endpoint for each kubernetes node.

Read more on prometheus auto discovery here:
- https://prometheus.io/docs/guides/file-sd/
- https://prometheus.io/docs/prometheus/latest/configuration/configuration/#file_sd_config

# Limitations
Currently dynamically discovering new nodes after prometheus has started is not possible because once the 
init container finished there is now way to get notified about new nodes. However, prometheus supports updating 
the `targets.json` at runtime, so the program could be extended to query the kubernetes api periodically or get
notified when nodes join or leave the cluster (not sure if the kubernetes api supports listening for changes) and 
this container could be run as a sidekick container within the prometheus pod to periodically update the config.
Because for my use case changes in the nodes are very range I will likely not add this feature, but if you need
this feature feel free to submit a pull request.

Also currently all path and settings are hard coded. This means the `targets.json` will always be written to `/var/prometheus/targets.json`.
This will likely be configurable via command line args or environment vars in the next version.

# Usage
You need a configmap that holds the `prometheus.yaml` that looks like this:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: monitoring
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/part-of: monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval:     15s # Set the scrape interval to every 15 seconds. Default is every 1 minute.
      evaluation_interval: 15s # Evaluate rules every 15 seconds. The default is every 1 minute.

    scrape_configs:
    - job_name: 'prometheus'
      static_configs:
        - targets: ['localhost:9090']
    - job_name: 'kube-metrics'
      scheme: https
      bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
      tls_config:
        ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
      file_sd_configs:
        - files:
          - '/var/prometheus/targets.json'
      relabel_configs:
        - source_labels: [__address__]
          regex:  '[^/]+(/.*)'            # capture '/...' part
          target_label: __metrics_path__  # change metrics path
        - source_labels: [__address__]
          regex:  '([^/]+)/.*'            # capture host:port
          target_label: __address__       # change target
```

and a deployment that looks like this:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
  labels:
    app.kubernetes.io/name: prometheus
    app.kubernetes.io/part-of: monitoring
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: prometheus
  template:
    metadata:
      labels:
        app.kubernetes.io/name: prometheus
        app.kubernetes.io/part-of: monitoring
        app.kubernetes.io/version: 1.0.11
    spec:
      serviceAccount: prometheus
      serviceAccountName: prometheus
      initContainers:
        - name: init
          image: danielr1996/prometheus-sd-kubernetes-node-metrics:v1.0.0
          volumeMounts:
            - name: config-dir
              mountPath: /var/prometheus
      containers:
        - name: prometheus
          image: prom/prometheus:v2.15.1
          ports:
            - containerPort: 9090
              name: http
              protocol: TCP
          volumeMounts:
            - name: config-dir
              mountPath: /var/prometheus
            - name: config-volume
              mountPath: /etc/prometheus/prometheus.yml
              subPath: prometheus.yml
      volumes:
        - name: config-volume
          configMap:
            name: prometheus-config
        - name: config-dir
          emptyDir: {}
---
```
