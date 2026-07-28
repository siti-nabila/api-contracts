PROTO_SRC := $(shell find proto -name '*.proto')
PROTO_OUT := pb

.PHONY: proto clean-proto test

proto:
	@echo "🔧 Generating protobuf contracts..."
	@mkdir -p $(PROTO_OUT)
	@for file in $(PROTO_SRC); do \
		echo "⏳ Generating $$file..."; \
		protoc \
			--proto_path=proto \
			--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
			--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
			$$file || exit 1; \
	done
	@echo "✅ Protobuf contracts generated in $(PROTO_OUT)/"

clean-proto:
	@echo "🧹 Cleaning generated protobuf files..."
	@find $(PROTO_OUT) -type f -name '*.pb.go' -delete

test:
	go test ./...
