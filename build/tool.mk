# ==============================================================================
# Makefile helper functions for tool
#

SHELL := /bin/bash
PROTOC := protoc

ifeq ($(origin ROOT_DIR),undefined)
	ROOT_DIR := $(abspath $(shell pwd -P))
endif


ifeq ($(origin PROTOC_GEN_PATH), undefined)
	PROTOC_GEN_PATH := $(ROOT_DIR)/tools/proto
endif

PROTOC_EXIST := $(shell type $(PROTOC) >/dev/null 2>&1 || { echo >&1 "not installed"; })

OS := $(word 1,$(subst _, ,$(PLATFORM)))
ARCH := $(word 2,$(subst _, ,$(PLATFORM)))

.PHONY: protoc.verify
protoc.verify:
ifneq (${PROTOC_EXIST},)
	$(error Please install protoc first)
endif
	@echo "=====> protoc verification passed <====="

.PHONY: proto.bin.verify
proto.bin.verify: protoc.verify
	@echo "====> verify protoc exist <===="
	$(if $(wildcard $(PROTOC_GEN_PATH)/bin/protoc-gen-go),,go get github.com/golang/protobuf/protoc-gen-go@v1.5.2 && go build -o $(PROTOC_GEN_PATH)/bin/protoc-gen-go github.com/golang/protobuf/protoc-gen-go)

.PHONY: proto.gen
proto.gen: proto.bin.verify 
	@echo "=====> generate grpc source codes <====="
	@sh $(ROOT_DIR)/script/proto.sh
