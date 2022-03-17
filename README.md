# cert-manager-webhook-yandex

Cert-manager ACME DNS webhook provider for Yandex Cloud.

## Installing

To install with helm, run:

```bash
$ helm repo add dodopizza https://dodopizza.github.io/cert-manager-webhook-yandex
$ helm repo update
$ helm install --name cert-manager-webhook-yandex dodopizza/cert-manager-webhook-yandex
```

OR

```bash
$ git clone https://github.com/dodopizza/cert-manager-webhook-yandex.git
$ cd cert-manager-webhook-yandex/deploy/cert-manager-webhook-yandex
$ helm install --name cert-manager-webhook-yandex .
```

### Issuer/ClusterIssuer

Get api key for service account with `dns.editor` permissions:

```bash
yc iam api-key create --service-account-name <service-account> --folder-id <folder-id>
```

An example issuer:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: yandex-authorized-key
type: Opaque
stringData:
  key: authorized-key-for-service-account
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-staging
  namespace: default
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: certmaster@gmail.com
    privateKeySecretRef:
      name: letsencrypt-staging-account-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.yandex.ru
          solverName: yandex
          config:
            apiKeySecretRef:
              name: yandex-authorized-key
              key: key

            folderId: <folder id where dns zone exists>

            # one of supported authorization types: iam-token or iam-key
            # this options depends on supplied secret
            # if oauth token specified, then value must be equal to `iam-token`
            # if authorized key for service account specified, then value must be equal to `iam-key`
            authorizationType: iam-key

            # optional field for dns challenge record ttl
            dnsRecordSetTTL: 120
```

And then you can issue a cert:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: tls-example-com
  namespace: default
spec:
  secretName: tls-example-com
  commonName: example.com
  issuerRef:
    name: letsencrypt-staging
    kind: Issuer
  dnsNames:
  - example.com
  - www.example.com
```

## Development

### Running the test suite

You can run the test suite with:

1. Generate api-key or oauth token key via cli or portal
2. Fill in the appropriate values in `testdata/yandex/credentials.yml` and `testdata/yandex/config.json`

```bash
$ TEST_ZONE_NAME=<set here dns zone>. make test-integration
```
