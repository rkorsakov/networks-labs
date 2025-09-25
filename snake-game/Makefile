PROTO_PATH=internal/proto
PROTO_FILE_PATH=internal/proto/snakes.proto
PROTO_GEN_PATH=internal/proto/gen

gen:
	@mkdir $(PROTO_GEN_PATH)
	@protoc -I $(PROTO_PATH) $(PROTO_FILE_PATH) \
	--go_out=$(PROTO_GEN_PATH) \
	--go_opt=paths=source_relative \
	--go-grpc_out=$(PROTO_GEN_PATH) \
	--go-grpc_opt=paths=source_relative
	@echo Successful generated

clean:
	@rm -rf $(PROTO_GEN_PATH)
	@echo Done!