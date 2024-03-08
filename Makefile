pb_gen:
	mkdir -p ./pb_autogen
	protoc --proto_path=proto --go_out=./pb_autogen --go_opt=paths=source_relative \
		--go-grpc_out=./pb_autogen --go-grpc_opt=paths=source_relative proto/*.proto

reg:
	cd registry && go build -o ../.local/registry *.go

srv:
	cd service && go build -o ../.local/service *.go

pb_gen_py:
	rm -rf ./service-py/pb
	cd ./service-py && python -m grpc_tools.protoc -Ipb=../proto --python_out=. \
		--pyi_out=. --grpc_python_out=. ../proto/*.proto