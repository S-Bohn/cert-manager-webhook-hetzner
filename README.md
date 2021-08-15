# ACME webhook for Hetzner DNS API

This is a webhook to be used with [cert-manager](https://www.cert-manager.io). It implements the DNS01 challenge solving logic with the [Hetzner DNS API](https://dns.hetzner.com/api-docs)

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Hetzner Webhook

```bash
helm repo add cert-manager-webhook-hetzner <https://helm.sbohn.dev/cert-manager-webhook-hetzner>
# Replace the groupName value with your desired domain
helm install --namespace cert-manager cert-manager-webhook-hetzner cert-manager-webhook-hetzner/cert-manager-webhook-hetzner --set groupName=acme.yourdomain.tld
```

## Configuration

Create a `ClusterIssuer` or `Issuer` resource as following:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: mail@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            # This group needs to be configured when installing the helm package, otherwise the webhook won't have permission to create an ACME challenge for this API group.
            groupName: acme.yourdomain.tld
            solverName: hetzner
            config:
              # optional: specify the key to use
              apiKeySecretRef: 
                name: hetzner-secret
                key: api-key
              # optional: specify the zone id, useful in combination with a filter to avoid zone lookup
              zoneId: razbZePHbywsVQRQmKzbdm
```

### Credentials
In order to access the Hetzner API, the webhook needs an API token.

If you choose another name for the secret than `hetzner-secret`, ensure you modify the value of `apiKeySecretRef.name` in the `[Cluster]Issuer` or adapt the default.

The secret for the example above will look like this:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hetzner-secret
type: Opaque
data:
  api-key: your-key-base64-encoded
```
