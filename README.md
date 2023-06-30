# Sample Kubernetes Controller using Kubebuilder

```bash
kubebuilder init --domain samples-k8s-ctrl.github.com  --repo github.com/etesami/sample-k8s-ctrl
kubebuilder create api --group calculator --version v1alpha1 --kind Calculator
```