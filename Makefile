
# import tool makefile
include build/tool.mk

# protoc to generate go.pb file
.PHONY: proto
proto:
	@$(MAKE) proto.gen