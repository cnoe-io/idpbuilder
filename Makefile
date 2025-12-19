LD_FLAGS=-ldflags " \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.idpbuilderVersion=$(shell git describe --always --tags --dirty --broken) \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.gitCommit=$(shell git rev-parse HEAD) \
    -X github.com/cnoe-io/idpbuilder/pkg/cmd/version.buildDate=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ') \
    "

# The name of the binary. Defaults to idpbuilder
OUT_FILE ?= idpbuilder

.PHONY: build
build: manifests generate fmt vet embedded-resources
	go build $(LD_FLAGS) -o $(OUT_FILE) main.go

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.1

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
KUSTOMIZE ?= $(LOCALBIN)/kustomize
HELM_TGZ ?= $(LOCALBIN)/helm.tar.gz
HELM ?= $(LOCALBIN)/helm

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.16.5
KUSTOMIZE_VERSION ?= v5.5.0

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
ifeq ($(RUN),)
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -p 1 --tags=integration ./... -coverprofile cover.out
else
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -p 1 --tags=integration ./... -coverprofile cover.out -run $(RUN)
endif

	

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./api/..."

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=pkg/controllers/resources

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

helm_os := $(shell uname | tr '[:upper:]' '[:lower:]')
helm_version ?= 3.15.0
ifeq ($(shell uname -m), x86_64)
	helm_arch ?= amd64
endif
ifeq ($(shell uname -m), arm64)
	helm_arch ?= arm64
endif
ifeq ($(shell uname -m), aarch64)
	helm_arch ?= arm64
endif


.PHONY: helm
helm: ## Download helm if necessary or use system helm
ifeq (,$(wildcard $(HELM)))
	@if command -v helm >/dev/null 2>&1; then \
		echo "Using system helm"; \
		ln -sf $$(command -v helm) $(HELM); \
	else \
		echo "Downloading helm v$(helm_version)"; \
		curl https://get.helm.sh/helm-v$(helm_version)-$(helm_os)-$(helm_arch).tar.gz -o $(HELM_TGZ); \
		tar xvzf $(HELM_TGZ) -C $(LOCALBIN) --strip-components 1 $(helm_os)-$(helm_arch)/helm; \
		chmod +x $(HELM); \
	fi
endif

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: embedded-resources
embedded-resources: kustomize helm
	export PATH=$(LOCALBIN):$$PATH; ./hack/embedded-resources.sh;

.PHONY: e2e
e2e: build
	go test -v -p 1 -timeout 15m --tags=e2e ./tests/e2e/...
