VERSION ?= latest

# set default data directory by OS
ifeq ($(OS),Windows_NT)
  DEFAULT_DATA_DIR=$(HOME)/.ledger
else
  ifeq ($(shell uname -s),Darwin)
    DEFAULT_DATA_DIR=$(HOME)/Library/ledger
  else
    DEFAULT_DATA_DIR=$(HOME)/.ledger
  endif
endif

GDBL_DATA_DIR ?= $(DEFAULT_DATA_DIR)

GODIST = ./build/dist
GO ?= latest
GORUN = env GO111MODULE=on go run

xtarget = $(strip $(subst build-,,$@)) # e.g. 'build-linux-amd64' -> 'linux-amd64'
xdest = $(GODIST)/$(xtarget)

# 'default' target builds all binaries for local development/testing
default: build-native

# 'release' target builds os-specific builds of only godbledger using xgo/docker
release: build-cross

clean:
	rm -rf build/.cache
	rm -rf build/bin
	rm -rf build/dist
	rm -rf release/
	rm -rf cert/

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*/*.proto

build-native:
	$(GORUN) utils/ci.go build

lint:
	$(GORUN) utils/ci.go lint

# our tests include an integration test which expects the local
# GOOS-based build output to be in the ./build/bin folder
test: build-native
	$(GORUN) utils/ci.go test

travis: build-native
	$(GORUN) utils/ci.go test -coverage $$TEST_PACKAGES

# -------------------------------------
# release_pattern=current
#
linux:
		mkdir -p release/godbledger-linux-x64-v$(VERSION)/
		GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o release/godbledger-linux-x64-v$(VERSION)/ ./...

linux-arm-7:
		mkdir -p release/godbledger-arm7-v$(VERSION)/
		env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on go build -o release/godbledger-arm7-v$(VERSION)/ ./...

linux-arm-64:
		mkdir -p release/godbledger-arm64-v$(VERSION)/
		env CC=aarch64-linux-gnu-gcc CXX=aarch-linux-gnu-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -o release/godbledger-arm64-v$(VERSION)/ ./...

# -------------------------------
# docker

# convenience target which looks like the other top-level build-* targets
build-docker: docker-build

docker-build:
	docker build -t godbledger:$(VERSION) -f ./Dockerfile.build .

docker-start:
	docker run -d --name=godbledger-server -p 50051:50051 --mount type=bind,source="$(GDBL_DATA_DIR)",target="/mnt/host/.ledger" --env GDBL_LOG_LEVEL=debug --env GDBL_DATA_DIR=/mnt/host/.ledger --env GDBL_CONFIG_FILE=/mnt/host/.ledger/config.docker.toml godbledger:$(VERSION) && docker logs godbledger-server

docker-stop:
	docker stop godbledger-server

docker-inspect:
	docker inspect godbledger-server

docker-clean:
	@$(if $(strip $(shell docker container list -a | grep godbledger-server)), @docker rm -f godbledger-server && echo "godbledger-server has been removed from docker",@echo "no godbledger-server container found")

docker-status:
	@$(if $(strip $(shell docker ps | grep godbledger-server)), @echo "godbledger-server is running on localhost:50051", @echo "godbledger-server is not running")

docker-logs:
	@docker logs godbledger-server

docker-logs-follow:
	@docker logs -f godbledger-server

docker-login:
	@$(if $(strip $(shell docker ps | grep godbledger-server)), @docker exec -it godbledger-server /bin/ash || 0, @docker run -it --rm --entrypoint /bin/ash godbledger:$(VERSION) )

# -------------------------------
# cross

build-cross: build-linux build-darwin build-windows

build-linux: build-linux-386 build-linux-amd64 build-linux-arm
	@echo "Linux cross compilation done:"
	@ls -ld $(GODIST)/linux-*

build-linux-386:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/386 -- --go=$(GO)

build-linux-amd64:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/amd64 -- --go=$(GO)

build-linux-arm: build-linux-arm-5 build-linux-arm-6 build-linux-arm-7 build-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GODIST)/linux-arm*

build-linux-arm-5:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/arm-5 -- --go=$(GO)

build-linux-arm-6:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/arm-6 -- --go=$(GO)

build-linux-arm-7:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/arm-7 -- --go=$(GO)

build-linux-arm64:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target linux/arm64 -- --go=$(GO)

build-darwin: build-darwin-10.6-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GODIST)/darwin-*

build-darwin-10.6-amd64:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target darwin-10.6/amd64 -- --go=$(GO)

build-windows: build-windows-4.0-386 build-windows-4.0-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GODIST)/windows-*

build-windows-4.0-386:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target windows-4.0/386 -- --go=$(GO)

build-windows-4.0-amd64:
	@echo "building $(xtarget)"
	$(GORUN) utils/ci.go xgo --target windows-4.0/amd64 -- --go=$(GO)

.PHONY: cert
cert:
	mkdir -p cert/
	cd cert; ../utils/gen.sh; cd ..
