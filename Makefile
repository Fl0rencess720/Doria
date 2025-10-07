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
		@echo "  user     - Run user service"
	@echo "  tts     - Run tts service"
	@echo "  mate     - Run mate service"
	@echo "  memory     - Run memory service"
	@echo ""
	@echo "Examples:"
	@echo "  make run gateway"
	@echo "  make run image"
		@echo "  make run user"
	@echo "  make run tts"
	@echo "  make run mate"
	@echo "  make run memory"

.PHONY: gateway image user tts mate memory
gateway:
	@go run src/gateway/cmd/main.go src/gateway/cmd/wire_gen.go --config src/gateway/configs

image:
	@go run src/services/image/cmd/main.go src/services/image/cmd/wire_gen.go --config src/services/image/configs

user:
	@go run src/services/user/cmd/main.go src/services/user/cmd/wire_gen.go --config src/services/user/configs
tts:
	@go run src/services/tts/cmd/main.go src/services/tts/cmd/wire_gen.go --config src/services/tts/configs
mate:
	@go run src/services/mate/cmd/main.go src/services/mate/cmd/wire_gen.go --config src/services/mate/configs
memory:
	@go run src/services/memory/cmd/main.go src/services/memory/cmd/wire_gen.go --config src/services/memory/configs