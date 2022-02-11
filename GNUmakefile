PLUGIN_BINARY=vf-device
export GO111MODULE=on

default: build

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf vf-device launcher

build:
	go build -o ${NOMAD_DEVICE_PLUGIN_DIR}/vf-plugin .

.PHONY: eval
eval: deps build
	./launcher device ${NOMAD_DEVICE_PLUGIN_DIR}/vf-plugin ./examples/config.hcl

.PHONY: fmt
fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -s -w ./...

.PHONY: bootstrap
bootstrap: deps # install all dependencies

.PHONY: launcher
deps:  ## Install build and development dependencies
	@echo "==> Updating build dependencies..."
	go build github.com/hashicorp/nomad/plugins/shared/cmd/launcher
