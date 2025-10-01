# Cert-Manager Integration for Secure Metrics (HTTPS)

This directory contains cert-manager resources for enabling HTTPS on the controller's metrics endpoint.

## Default Configuration (HTTP)

**By default, metrics are served on HTTP (port 8080)** to avoid x509 certificate errors when cert-manager is not installed. This is suitable for most deployments where metrics scraping happens within a trusted network.

## Why Enable HTTPS with Cert-Manager?

For production environments requiring encrypted metrics, you can enable HTTPS with cert-manager. This will:

1. **Serve metrics on port 8443 with TLS**
2. **Generate valid certificates** with the correct DNS names as Subject Alternative Names (SANs)
3. **Enable secure metrics scraping** from Prometheus or other monitoring tools

Without cert-manager, if you manually enable HTTPS (`--metrics-secure=true`), controller-runtime generates self-signed certificates with only `localhost/127.0.0.1` as SANs, which causes x509 errors when scrapers use the service DNS name.

## Prerequisites

Install cert-manager in your cluster:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml
```

## Enabling Cert-Manager Integration

1. **Edit `config/default/kustomization.yaml`** and uncomment the following sections:

   ```yaml
   # Line 25 - Add cert-manager resources
   resources:
   - ../certmanager  # Uncomment this line

   # Lines 47-49 - Apply certificate patch
   patches:
   - path: cert_metrics_manager_patch.yaml  # Uncomment this block
     target:
       kind: Deployment

   # Lines 59-118 - Enable kustomize replacements for DNS injection
   replacements:  # Uncomment the entire replacements section
   - source:
       kind: Service
       ...
   ```

2. **Regenerate manifests**:

   ```bash
   cd pkg/k8s/crd
   make generate-controller-yaml generate-install-yaml
   ```

3. **Deploy**:

   ```bash
   kubectl apply -f install/shifu_install.yml
   ```

4. **Verify certificate creation**:

   ```bash
   kubectl get certificate -n shifu-crd-system
   kubectl get secret metrics-server-cert -n shifu-crd-system
   kubectl describe certificate shifu-crd-metrics-certs -n shifu-crd-system
   ```

## What Gets Created

- **Issuer**: `shifu-crd-selfsigned-issuer` - A self-signed certificate issuer
- **Certificate**: `shifu-crd-metrics-certs` - Certificate with proper DNS SANs:
  - `shifu-crd-controller-manager-metrics-service.shifu-crd-system.svc`
  - `shifu-crd-controller-manager-metrics-service.shifu-crd-system.svc.cluster.local`
- **Secret**: `metrics-server-cert` - Contains `ca.crt`, `tls.crt`, and `tls.key`

The certificate is automatically mounted to `/tmp/k8s-metrics-server/metrics-certs` in the controller pod.

## Disabling (Default Behavior)

To disable cert-manager integration and use controller-runtime's self-signed certificates:

1. Comment out the sections mentioned above in `config/default/kustomization.yaml`
2. Regenerate manifests: `make generate-controller-yaml generate-install-yaml`

**Note**: This is the default configuration. The controller will still work but Prometheus scraping will fail with x509 errors unless you configure Prometheus with `insecure_skip_verify: true` (not recommended for production).
