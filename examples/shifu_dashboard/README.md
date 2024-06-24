# Shifu Dashboard

Shifu Dashboard is a monitoring tool for Shifu based on [Grafana](https://grafana.com/) and [Prometheus](https://prometheus.io/).
It can be used to monitor the performance of Shifu in real-time, like memory usage, CPU usage, current deviceShifu num and so on.

## Requirements

- Kubernetes cluster
- Helm
- Shifu

## Installation

Before start, you need to make sure your cluster is ready and shifu had been installed.

1. Install Prometheus-Grafana stack by helm
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts && \
helm repo update && \
helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
```

2. Import the dashboard to Grafana
After Prometheus-Grafana stack installed, you can access the Grafana dashboard by port-forwarding the service to your local machine.
```bas
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
```
Then you can access the Grafana dashboard by visiting `http://localhost:3000` in your browser. The default username is `admin` and the password is `prom-operator`.

After you login, you can import the dashboard by using the following steps:
- Click the `+` icon on the left side of the dashboard list
- Click `New` - `Import` 
- Upload the `shifu_dashboard.json` file in this folder
- Choose the Prometheus data source
- Click `Import`
- Now you can see the Shifu dashboard in the dashboard list
