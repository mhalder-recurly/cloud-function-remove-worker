module github.com/mhalder-recurly/cloud-function-remove-worker

go 1.16

require (
	github.com/spf13/viper v1.7.1
	google.golang.org/api v0.13.0
	k8s.io/client-go v0.0.0-00010101000000-000000000000
)

replace (
	k8s.io/api => k8s.io/kubernetes/staging/src/k8s.io/api v0.0.0-20200813160325-9f2892aab98f
	k8s.io/apimachinery => k8s.io/kubernetes/staging/src/k8s.io/apimachinery v0.0.0-20200813160325-9f2892aab98f
	k8s.io/client-go => k8s.io/kubernetes/staging/src/k8s.io/client-go v0.0.0-20200813160325-9f2892aab98f
)
