SHELL = /bin/bash
REG = docker.io
ORG = odranoel
IMAGE = deployment-operator
TAG = latest
SA = deployment-operator
NS = deployment-operator
TEST_FOLDER = ./test/e2e

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

image/build-with-tests:
	@operator-sdk build --enable-tests ${REG}/${ORG}/${IMAGE}:${TAG}

image/push:
	@docker push ${REG}/${ORG}/${IMAGE}:${TAG}

test/unit:
	@go test -v -race -cover ./pkg/...

test/smoke: code/check test/unit

test/e2e/local: image/build-with-tests image/push
	@operator-sdk test local ${TEST_FOLDER} --go-test-flags "-v"

test/e2e/cluster: image/build-with-tests image/push
	@kubectl apply -f deploy/test-pod.yaml -n ${NS}
	${SHELL} ./scripts/stream-pod ${TEST_POD_NAME} ${NS}

cluster/prepare:
	@kubectl create ns ${NS} || true
	@kubectl apply -f deploy/role.yaml -n ${NS}
	@kubectl apply -f deploy/service_account.yaml -n ${NS}
	@kubectl apply -f deploy/role_binding.yaml -n ${NS}
	@kubectl apply -f deploy/crds/integreatly_v1alpha1_deployment_crd.yaml

cluster/deploy:
	@kubectl apply -f deploy/crds/integreatly_v1alpha1_deployment_cr.yaml -n ${NS}
	@kubectl apply -f deploy/operator.yaml -n ${NS}

cluster/clean:
	@kubectl delete all --all -n ${NS}
	kubectl delete tdeployment/example-deployment -n ${NS}
