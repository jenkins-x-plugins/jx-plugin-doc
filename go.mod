module github.com/jenkins-x-plugins/jx-plugin-doc

require (
	github.com/jenkins-x/go-scm v1.6.12
	github.com/jenkins-x/jx-helpers/v3 v3.0.92
	github.com/jenkins-x/jx-logging/v3 v3.0.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
)

replace (
	k8s.io/api => k8s.io/api v0.20.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.3
	k8s.io/client-go => k8s.io/client-go v0.20.3
)

go 1.15
