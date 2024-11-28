helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack  --version 45.7.1 --namespace monitoring --create-namespace
kubectl apply -f grafana-svc-lb.yml