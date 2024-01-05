# Sample Kubernetes Controller using Kubebuilder

```bash
kubebuilder init --domain samples-k8s-ctrl.github.com  --repo github.com/etesami/sample-k8s-ctrl
kubebuilder create api --group calculator --version v1alpha1 --kind Calculator
```

### api/v1alpha1/calculator_types.go
```go
// Omitted for brevity

type CalculatorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Calculator. Edit calculator_types.go to remove/update
	NumberOne int    `json:"numberone"`
	NumberTwo int    `json:"numbertwo"`
	Operation string `json:"operation"`
}

// CalculatorStatus defines the observed state of Calculator
type CalculatorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Result int `json:"result"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Number One",type="integer",JSONPath=".spec.numberone",description="Input number one"
//+kubebuilder:printcolumn:name="Number Two",type="integer",JSONPath=".spec.numbertwo",description="Input number two"
//+kubebuilder:printcolumn:name="Operation",type="string",JSONPath=".spec.operation",description="Operation"
//+kubebuilder:printcolumn:name="Result",type="integer",JSONPath=".status.result",description="Sum of two numbers"

// Omitted for brevity

```

### internal/controller/calculator_controller.go
```go
// Omitted for brevity

var calculator calculatorv1alpha1.Calculator
if err := r.Get(ctx, req.NamespacedName, &calculator); err != nil {
    klog.Errorf("unable to fetch Calculator: %v", err)
    return ctrl.Result{}, client.IgnoreNotFound(err)
}
klog.Infof("\nCalculator: %v\n", calculator)
switch calculator.Spec.Operation {
case "add":
    calculator.Status.Result = calculator.Spec.NumberOne + calculator.Spec.NumberTwo
case "subtract":
    calculator.Status.Result = calculator.Spec.NumberOne - calculator.Spec.NumberTwo
case "multiply":
    calculator.Status.Result = calculator.Spec.NumberOne * calculator.Spec.NumberTwo
case "divide":
    calculator.Status.Result = calculator.Spec.NumberOne / calculator.Spec.NumberTwo
default:
    klog.Errorf("unknown operation: %v", calculator.Spec.Operation)
}

if err := r.Status().Update(ctx, &calculator); err != nil {
    klog.Errorf("unable to update Calculator status: %v", err)
    return ctrl.Result{}, err
}

klog.Infof("Calculator status updated successfully: %v", calculator.Status.Result)

// Omitted for brevity

```

```bash
make manifest
make install
```
## Implementing a Validating Webhook
```bash
kubebuilder create webhook --group calculator --version v1alpha1 --kind Calculator --defaulting --programmatic-validation
```

### api/v1alpha1/calculator_webhook.go
```go
// Omitted for brevity

func (r *Calculator) ValidateCreate() (admission.Warnings, error) {
	calculatorlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	klog.Infof("Validate create", "name", r.Name)
	if !isInList([]string{"add", "subtract", "multiply", "divide"}, r.Spec.Operation) {
		return nil, fmt.Errorf("Operation %s is not supported", r.Spec.Operation)
	}
	return nil, nil
}

// Omitted for brevity

func isInList(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
```

Look [here](https://github.com/kubernetes-sigs/kubebuilder/pull/3456/files) for a bug in Kubebuilder version v3.11.0.

```bash
make manifest
make install

# To build and push the image to repository:
make docker-build docker-push IMG=etesami/sample-k8s-ctrl:latest

# To deply the controller as pod in the cluster
make deploy IMG=etesami/sample-k8s-ctrl:latest

# To undeploy
make undeploy
```

If you run the controller locally, you may use `export ENABLE_WEBHOOKS=false` and this will force the webhook to be diabled. 
Otherwise, you need to provide certificates. Follow instructions [here](https://book.kubebuilder.io/cronjob-tutorial/running-webhook)
to learn how to inject certificate using kastomize.

## Samples
```yml
apiVersion: calculator.samples-k8s-ctrl.github.com/v1alpha1
kind: Calculator
metadata:
  name: calculator-wrong1
spec:
  numberone: 10
  numbertwo: 20
  operation: "sth"
---
apiVersion: calculator.samples-k8s-ctrl.github.com/v1alpha1
kind: Calculator
metadata:
  name: calculator-wrong1
spec:
  numberone: 10
  numbertwo: 20
  operation: "add"
```

## Multi-Group APIs

After the first version, we moved to the multi-group APIs. We need to follow this [guide](https://kubebuilder.io/migration/multi-group.html) first and then we can create a new API under a new group. 

```bash
kubebuilder edit --multigroup=true
# Create calculator folder in api/ and internal/controller directories
# Move related files in each folder to the newly created folder
# Change all path every where related to api/v1alpha1 to api/calculator/v1alpha1
# Change all path every where related to internal/controller to internal/controller/calculator

# Create the new API under a new group
kubebuilder create api --group test1 --version v1alpha1 --kind SimpleDeployment
```
A SimpleDeployment resource creates a deployment and maintain its number of replica and pod information. Take a look at the `api/test1/v1alpha1/simpledeployment_types.go`:
```yaml
apiVersion: test1.samples-k8s-ctrl.github.com/v1alpha1
kind: SimpleDeployment
metadata:
  name: simpledeployment-sample
spec:
  # For custom field defined here see: api/test1/v1alpha1/simpledeployment_types.go
  # Custom field is defined
  replicas: 1
  # Custom field using K8S metav1.LabelSelector library
  selector:
    matchLabels:
      app: nginx
  # Custom field using K8S v1.PodTemplateSpec library
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
  # Custom field by defining a custom struct
  # I could not get the this working using metav1.ObjectMeta library
  mymetadata: 
    labels: 
      app: nginx
```
Also, take a look at the `internal/controller/test1/simpledeployment_controller.go`.