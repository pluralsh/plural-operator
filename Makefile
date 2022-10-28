
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

CRD_OPTIONS ?= "crd"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

generate: controller-gen generate-client ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations and the clientset, informers and listers.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate fmt vet ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.7.2/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

unit-test:
	go test -tags=unit -v -race ./controllers/...

##@ Build

build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

run: manifests generate fmt vet ## Run a controller from your host.
	ENABLE_WEBHOOKS=false go run ./main.go -zap-devel -zap-log-level 2

docker-build: build ## Build docker image with the manager.
	docker build -t ${IMG} .

docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -

run-client-gen:
	$(CLIENT_GEN) --clientset-name versioned --input-base github.com/pluralsh/plural-operator/apis --input platform/v1alpha1,vpn/v1alpha1 --output-package github.com/pluralsh/plural-operator/generated/client/clientset --go-header-file hack/boilerplate.go.txt

run-lister-gen:
	$(LISTER_GEN) --input-dirs github.com/pluralsh/plural-operator/apis/platform/v1alpha1,github.com/pluralsh/plural-operator/apis/vpn/v1alpha1 --output-package github.com/pluralsh/plural-operator/generated/client/listers --go-header-file hack/boilerplate.go.txt

run-informer-gen:
	$(INFORMER_GEN) --input-dirs github.com/pluralsh/plural-operator/apis/platform/v1alpha1,github.com/pluralsh/plural-operator/apis/vpn/v1alpha1  --versioned-clientset-package github.com/pluralsh/plural-operator/generated/client/clientset/versioned --listers-package github.com/pluralsh/plural-operator/generated/client/listers --output-package github.com/pluralsh/plural-operator/generated/client/informers --go-header-file hack/boilerplate.go.txt

generate-client: run-client-gen run-lister-gen run-informer-gen
	rm -rf generated
	mv github.com/pluralsh/plural-operator/generated generated
	rm -rf github.com

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.9.2)

KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize: ## Download kustomize locally if necessary.
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v4@v4.5.7)

CLIENT_GEN = $(shell pwd)/bin/client-gen
client-gen: ## Download client-gen locally if necessary.
	$(call go-get-tool,$(CLIENT_GEN),k8s.io/code-generator/cmd/client-gen@v0.25.3)

LISTER_GEN = $(shell pwd)/bin/lister-gen
lister-gen: ## Download lister-gen locally if necessary.
	$(call go-get-tool,$(LISTER_GEN),k8s.io/code-generator/cmd/lister-gen@v0.25.3)

INFORMER_GEN = $(shell pwd)/bin/informer-gen
informer-gen: ## Download informer-gen locally if necessary.
	$(call go-get-tool,$(INFORMER_GEN),k8s.io/code-generator/cmd/informer-gen@v0.25.3)

DEFFAULTER_GEN = $(shell pwd)/bin/defaulter-gen
defaulter-gen: ## Download defaulter-gen locally if necessary.
	$(call go-get-tool,$(DEFFAULTER_GEN),k8s.io/code-generator/cmd/defaulter-gen@v0.25.3)

DEEPCOPY_GEN = $(shell pwd)/bin/deepcopy-gen
deepcopy-gen: ## Download deepcopy-gen locally if necessary.
	$(call go-get-tool,$(DEEPCOPY_GEN),k8s.io/code-generator/cmd/deepcopy-gen@v0.25.3)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
