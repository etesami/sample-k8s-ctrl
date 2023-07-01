# Sample Kubernetes Controller using Kubebuilder

```bash
kubebuilder init --domain samples-k8s-ctrl.github.com  --repo github.com/etesami/sample-k8s-ctrl
kubebuilder create api --group calculator --version v1alpha1 --kind Calculator
```

## Implementing a Validating Webhook
```bash
kubebuilder create webhook --group calculator --version v1alpha1 --kind Calculator --defaulting --programmatic-validation
```