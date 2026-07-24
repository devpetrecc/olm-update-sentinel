IMG ?= ghcr.io/devpetrecc/olm-update-sentinel:latest

# Tooling paths
LOCALBIN ?= $(CURDIR)/bin
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
# Updated to v0.17.2 to support newer Go toolchains
CONTROLLER_TOOLS_VERSION ?= v0.17.2

.PHONY: all build test docker-build docker-push deploy manifests generate controller-gen

all: test build

test:
	go test ./... -coverprofile cover.out

build:
	go build -o bin/manager cmd/main.go

# Generate CustomResourceDefinitions, ClusterRole, and Webhooks
manifests: controller-gen
	"$(CONTROLLER_GEN)" rbac:roleName=manager-role crd paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations
generate: controller-gen
	"$(CONTROLLER_GEN)" object paths="./..."

docker-build:
	docker build -t $(IMG) .

docker-push:
	docker push $(IMG)

deploy: manifests
	kustomize build config/default | kubectl apply -f -

## Tool Dependencies
$(LOCALBIN):
	mkdir -p "$@"

controller-gen: $(LOCALBIN)
	@if [ ! -s "$(CONTROLLER_GEN)" ]; then \
		GOBIN="$(LOCALBIN)" go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION) ; \
	fi