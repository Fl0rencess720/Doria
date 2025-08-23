.PHONY: run

run:
	go run src/gateway/cmd/main.go src/gateway/cmd/wire_gen.go --config src/gateway/configs