IDL_DIR := src/idl
RPC_DIR := src/rpc
PROTO_FILES := $(wildcard $(IDL_DIR)/*.proto)
PROTO_GEN_DIRS := $(patsubst $(IDL_DIR)/%.proto,$(RPC_DIR)/%,$(basename $(notdir $(PROTO_FILES))))

.PHONY: proto
proto:
	@rm -rf $(PROTO_GEN_DIRS)
	@protoc -I $(IDL_DIR) --go_out=src --go-grpc_out=src $(PROTO_FILES)

.PHONY: run help
run: help

help:
	@echo "Usage: make run <target>"
	@echo "Available targets:"
	@echo "  gateway   - Run gateway service"
	@echo "  image     - Run image service"
	@echo "  chat     - Run chat service"
	@echo "  user     - Run user service"
	@echo ""
	@echo "Examples:"
	@echo "  make run gateway"
	@echo "  make run image"
	@echo "  make run chat"
	@echo "  make run user"

.PHONY: gateway image
gateway:
	@go run src/gateway/cmd/main.go src/gateway/cmd/wire_gen.go --config src/gateway/configs

image:
	@go run src/services/image/cmd/main.go src/services/image/cmd/wire_gen.go --config src/services/image/configs

chat:
	@go run src/services/chat/cmd/main.go src/services/chat/cmd/wire_gen.go --config src/services/chat/configs
user:
	@go run src/services/user/cmd/main.go src/services/user/cmd/wire_gen.go --config src/services/user/configs