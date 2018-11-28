SHELL = /bin/bash
REG = docker.io
ORG = odranoel
IMAGE = deployment-operator
TAG = latest
SA = deployment-operator
NS = deployment-operator

setup/dep:
	@dep ensure -v

code/gen:
	@operator-sdk generate k8s

code/check:
	@diff -u <(echo -n) <(gofmt -d `find . -type f -name '*.go' -not -path "./vendor/*"`)

code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

image/build:
	@operator-sdk build ${REG}/${ORG}/${IMAGE}:${TAG}

image/push:
	docker push ${REG}/${ORG}/${IMAGE}:${TAG}

cluster/prepare:
	@kubectl apply -f deploy/role.yaml -n ${NS}
	@kubectl apply -f deploy/service_account.yaml -n ${NS}
	@kubectl apply -f deploy/role_binding.yaml -n ${NS}
	@kubectl apply -f deploy/crds/integreatly_v1alpha1_deployment_crd.yaml

cluster/deploy:
	@kubectl apply -f deploy/crds/integreatly_v1alpha1_deployment_cr.yaml -n ${NS}
	@kubectl apply -f deploy/operator.yaml -n ${NS}
