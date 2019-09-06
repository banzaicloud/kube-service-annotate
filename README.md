# kube-service-annotate

Kubernetes mutating admission webhook to automatically annotate services.


Command line arguments:

- `--debug` *Log debug messages*
- `--listen-addr` string *Listen address (default ":8080")*
- `--rules-file` string *Rule file path (default "config.yaml")*
- `--tls-cert-file` string *TLS certificate file*
- `--tls-key-file` string *TLS key file*


Example rules-file

You can configure multiple annotation rules. Use kubernetes like selectors to filter services.

```yaml
- selector:
    app: http-service
  annotations:
    service.tls/enabled: true
- annotations:
    service.beta.kubernetes.io/azure-load-balancer-internal: "true"
```

Install with Helm

```
helm repo add banzaicloud-stable https://kubernetes-charts.banzaicloud.com/
helm install banzaicloud-stable/kube-service-annotate
```