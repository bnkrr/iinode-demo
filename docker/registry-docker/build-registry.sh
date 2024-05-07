
PROJECT_ROOT=../../

cd $PROJECT_ROOT # run at project root

# CGO_ENABLED=0 is needed, see https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
CGO_ENABLED=0 go build -o docker/registry-docker/registry registry/*.go
