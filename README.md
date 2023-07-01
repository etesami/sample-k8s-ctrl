# Sample Kubernetes Controller using Kubebuilder

```bash
kubebuilder init --domain samples-k8s-ctrl.github.com  --repo github.com/etesami/sample-k8s-ctrl
kubebuilder create api --group calculator --version v1alpha1 --kind Calculator
```

## Implementing a Validating Webhook
```bash
kubebuilder create webhook --group calculator --version v1alpha1 --kind Calculator --defaulting --programmatic-validation
```
After this, you would need to add keys for an already generated certificate, Look below:
```yml
# Create a namespace
apiVersion: v1
kind: Namespace
metadata:
  name: sandbox
---
# Cluster scoped issuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}
---
# Create a Certificate
# Assuming we created a cluster scoped issuer previously:
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: my-selfsigned-ca
  namespace: sandbox
spec:
  isCA: true
  commonName: my-selfsigned-ca
  secretName: root-secret
  privateKey:
    algorithm: ECDSA
    size: 256
  issuerRef:
    name: selfsigned-issuer
    kind: ClusterIssuer
    group: cert-manager.io
---
# Create a CA Issuer. This is a simple Issuer that will sign certificates based on a private key. 
# This issuer is not a ClusterIssuer but a regular Issuer.
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: my-ca-issuer
  namespace: sandbox
spec:
  ca:
    secretName: root-secret
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: webhook-certificate
  namespace: sandbox
spec:
  # Secret names are always required.
  secretName: webhook-tls
  commonName: webhook-service.sandbox.svc
  isCA: false
  privateKey:
    algorithm: RSA
    encoding: PKCS1
    size: 2048
  # Issuer references are always required.
  issuerRef:
    name: my-ca-issuer
    # We can reference ClusterIssuers by changing the kind here.
    # The default value is Issuer (i.e. a locally namespaced Issuer)
    kind: Issuer
```
and now:
```bash
# Run the following command for ca.crt, tls.crt, tls.key
kubectl get secret webhook-tls -n sandbox -o json | jq -r '.data."tls.key"'  | base64 --decode > /tmp/k8s-webhook-server/serving-certs/tls.key
```