# Image URL to use all building/pushing image targets
#REPO ?= 'ccr.ccs.tencentyun.com/ccs-dev'
REPO ?= ccr.ccs.tencentyun.com/chenky
TAG ?= 0.0.4
GENERIC_IMG=$(REPO)/clusterapi-generic-controller:$(TAG)
CLUSTER_IMG=$(REPO)/tke-cluster-controller:$(TAG)
MACHINE_IMG=$(REPO)/tke-machine-controller:$(TAG)

CLUSTER-CONTROLLER=output/bin/tke-cluster-controller
MACHINE-CONTROLLER=output/bin/tke-machine-controller
GENERIC-CONTROLLER=output/bin/clusterapi-generic-controller

GCFLAGS ?= -x -gcflags="-N -l"

all: bin 

# Run tests
test: generate fmt vet manifests
	go test -v -tags=integration ./pkg/... ./cmd/... -coverprofile cover.out

tke-cluster-controller: ${CLUSTER-CONTROLLER}

tke-machine-controller: ${MACHINE-CONTROLLER}

clusterapi-generic-controller: ${GENERIC-CONTROLLER}

${CLUSTER-CONTROLLER}:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' ${GCFLAGS} -o $@ cmd/tke-cluster-controller/main.go	

${MACHINE-CONTROLLER}:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' ${GCFLAGS} -o $@ cmd/tke-machine-controller/main.go

${GENERIC-CONTROLLER}:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' ${GCFLAGS} -o $@ cmd/clusterapi-generic-controller/main.go

bin: ${CLUSTER-CONTROLLER} ${MACHINE-CONTROLLER} ${GENERIC-CONTROLLER}

clean:
	rm -f output/bin/* output/yaml/*

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager sigs.k8s.io/cluster-api-provider-gcp/cmd/manager

# Build manager binary
clusterctl: generate fmt vet
	go build -o bin/clusterctl sigs.k8s.io/cluster-api-provider-gcp/cmd/clusterctl

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crds
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...


YAML_DIR=output/yaml
GENERIC_CTL_YAML=output/yaml/clusterapi-generic-controller.yaml
CLUSTER_CTL_YAML=output/yaml/tke-cluster-controller.yaml
MACHINE_CTL_YAML=output/yaml/tke-machine-controller.yaml
ALL_IN_ONE_YAML=output/yaml/clusterapi-controllers-all-in-one.yaml

${YAML_DIR}:
	mkdir -p $@

# Build the docker image 
docker-build-generic-controller: ${YAML_DIR}
	docker build -f cmd/clusterapi-generic-controller/Dockerfile  . -t ${GENERIC_IMG}

docker-build-tke-cluster-controller: ${YAML_DIR}
	docker build -f cmd/tke-cluster-controller/Dockerfile  . -t ${CLUSTER_IMG}

docker-build-tke-machine-controller: ${YAML_DIR}
	docker build -f cmd/tke-machine-controller/Dockerfile  . -t ${MACHINE_IMG}

# Build the docker image from local biniaries in output/bin
img-generic-controller: ${GENERIC-CONTROLLER}
	docker build -f cmd/clusterapi-generic-controller/Dockerfile.local . -t ${GENERIC_IMG}

img-cluster-controller: ${CLUSTER-CONTROLLER}
	docker build -f cmd/tke-cluster-controller/Dockerfile.local . -t ${CLUSTER_IMG}

img-machine-controller: ${MACHINE-CONTROLLER}
	docker build -f cmd/tke-machine-controller/Dockerfile.local . -t ${MACHINE_IMG}

${GENERIC_CTL_YAML}: ${YAML_DIR}
	sed -e 's@image: .*@image: '"${GENERIC_IMG}"'@' ./config/controller/clusterapi-generic-controller.yaml > $@

${CLUSTER_CTL_YAML}: ${YAML_DIR}
	sed -e 's@image: .*@image: '"${CLUSTER_IMG}"'@' ./config/controller/tke-cluster-controller.yaml > $@

${MACHINE_CTL_YAML}: ${YAML_DIR}
	sed -e 's@image: .*@image: '"${MACHINE_IMG}"'@' ./config/controller/tke-machine-controller.yaml > $@
	

yaml: 
	cp config/controller/tencent-cloud-api-secret.yaml  ${YAML_DIR}/
	make ${GENERIC_CTL_YAML}
	make ${CLUSTER_CTL_YAML}
	make ${MACHINE_CTL_YAML}
	cat config/rbac/rbac_role.yaml >> ${ALL_IN_ONE_YAML}
	echo "---" >> ${ALL_IN_ONE_YAML}
	cat config/rbac/rbac_role_binding.yaml >> ${ALL_IN_ONE_YAML}
	echo "---" >> ${ALL_IN_ONE_YAML}
	cat ${GENERIC_CTL_YAML}	>> ${ALL_IN_ONE_YAML}
	echo "---" >> ${ALL_IN_ONE_YAML}
	cat ${CLUSTER_CTL_YAML} >> ${ALL_IN_ONE_YAML}
	echo "---" >> ${ALL_IN_ONE_YAML}
	cat ${MACHINE_CTL_YAML} >> ${ALL_IN_ONE_YAML}


img-by-docker: docker-build-generic-controller docker-build-tke-cluster-controller docker-build-tke-machine-controller yaml

img: img-generic-controller img-cluster-controller img-machine-controller

push:
	docker push ${GENERIC_IMG}
	docker push ${CLUSTER_IMG}
	docker push ${MACHINE_IMG}

help:
	@echo '## build all binaries of controllers'
	@echo 'make bin'
	@echo 'make bin/clusterapi-generic-controller'
	@echo 'make bin/tke-cluster-controller'
	@echo 'make bin/tke-machine-controller'
	@echo '## build all docker images from local binaries'
	@echo 'REPO=ccr.ccs.tencentyun.com/ccs-dev TAG=0.2 make img'
	@echo 'make img-cluster-controller'
	@echo 'make docker-build-generic-controller'
	@echo 'make docker-build-tke-cluster-controller'
	@echo 'make docker-build-tke-machine-controller'
	@echo '## push images'
	@echo 'REPO=ccr.ccs.tencentyun.com/ccs-dev TAG=0.2 make push'

.PHONY : help clean controllers all bin img push img-by-docker img-generic-controller img-cluster-controller img-machine-controller  tke-cluster-controller tke-machine-controller clusterapi-generic-controller docker-build-generic-controller docker-build-tke-cluster-controller docker-build-tke-machine-controller
