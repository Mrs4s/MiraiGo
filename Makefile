PROTO_DIR=client/pb
PROTO_OUTPUT_PATH=client
PROTO_IMPORT_PATH=client

PROTO_FILES := \
	$(PROTO_DIR)/*.proto \
	$(PROTO_DIR)/channel/*.proto  \
	$(PROTO_DIR)/cmd0x3f6/*.proto \
	$(PROTO_DIR)/cmd0x6ff/*.proto \
	$(PROTO_DIR)/cmd0x346/*.proto \
	$(PROTO_DIR)/cmd0x352/*.proto \
	$(PROTO_DIR)/cmd0x388/*.proto \
	$(PROTO_DIR)/exciting/*.proto \
	$(PROTO_DIR)/faceroam/*.proto \
	$(PROTO_DIR)/highway/*.proto  \
	$(PROTO_DIR)/longmsg/*.proto  \
	$(PROTO_DIR)/msf/*.proto      \
	$(PROTO_DIR)/msg/*.proto      \
	$(PROTO_DIR)/msgtype0x210/*.proto \
	$(PROTO_DIR)/multimsg/*.proto     \
	$(PROTO_DIR)/notify/*.proto       \
	$(PROTO_DIR)/oidb/*.proto         \
	$(PROTO_DIR)/profilecard/*.proto  \
	$(PROTO_DIR)/pttcenter/*.proto    \
	$(PROTO_DIR)/qweb/*.proto         \
	$(PROTO_DIR)/richmedia/*.proto    \
	$(PROTO_DIR)/structmsg/*.proto    \
	$(PROTO_DIR)/web/*.proto

PROTOC_GEN_GOLITE_VERSION := \
	$(shell grep "github.com/RomiChan/protobuf" go.mod | awk -F v '{print "v"$$2}')

.PHONY: protoc-gen-golite-version clean install-protoc-plugin proto
.DEFAULT_GOAL := proto

protoc-gen-golite-version:
	@echo "Use protoc-gen-golite version: $(PROTOC_GEN_GOLITE_VERSION)"

clean:
	find . -name "*.pb.go" | xargs rm -f

install-protoc-plugin: protoc-gen-golite-version
	go install github.com/RomiChan/protobuf/cmd/protoc-gen-golite@$(PROTOC_GEN_GOLITE_VERSION)

proto: install-protoc-plugin
	protoc --golite_out=$(PROTO_OUTPUT_PATH) --golite_opt=paths=source_relative -I=$(PROTO_IMPORT_PATH) $(PROTO_FILES)
