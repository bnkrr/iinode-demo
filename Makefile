pb_gen:
	rm -rf ./pb_autogen/go.mod
	mkdir -p ./pb_autogen
	protoc --proto_path=proto --go_out=./pb_autogen --go_opt=paths=source_relative \
		--go-grpc_out=./pb_autogen --go-grpc_opt=paths=source_relative proto/*.proto
	cd pb_autogen && go mod init github.com/bnkrr/iinode-demo/proto && go mod tidy

reg:
	cd registry && go build -o ../.local/registry *.go

srv:
	cd service && go build -o ../.local/service *.go
