# Auto generated binary variables helper managed by https://github.com/bwplotka/bingo v0.3.1. DO NOT EDIT.
# All tools are designed to be build inside $GOBIN.
BINGO_DIR := $(dir $(lastword $(MAKEFILE_LIST)))
GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(firstword $(subst :, ,${GOPATH}))/bin
GO     ?= $(shell which go)

# Bellow generated variables ensure that every time a tool under each variable is invoked, the correct version
# will be used; reinstalling only if needed.
# For example for kustomize variable:
#
# In your main Makefile (for non array binaries):
#
#include .bingo/Variables.mk # Assuming -dir was set to .bingo .
#
#command: $(KUSTOMIZE)
#	@echo "Running kustomize"
#	@$(KUSTOMIZE) <flags/args..>
#
KUSTOMIZE := $(GOBIN)/kustomize-v3.8.7
$(KUSTOMIZE): $(BINGO_DIR)/kustomize.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/kustomize-v3.8.7"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=kustomize.mod -o=$(GOBIN)/kustomize-v3.8.7 "sigs.k8s.io/kustomize/kustomize/v3"

OPERATOR_SDK := $(GOBIN)/operator-sdk-v1.3.0
$(OPERATOR_SDK): $(BINGO_DIR)/operator-sdk.mod
	@# Install binary/ries using Go 1.14+ build command. This is using bwplotka/bingo-controlled, separate go module with pinned dependencies.
	@echo "(re)installing $(GOBIN)/operator-sdk-v1.3.0"
	@cd $(BINGO_DIR) && $(GO) build -mod=mod -modfile=operator-sdk.mod -o=$(GOBIN)/operator-sdk-v1.3.0 "github.com/operator-framework/operator-sdk/cmd/operator-sdk"

