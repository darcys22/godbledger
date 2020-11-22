BINARY := godbledger
VERSION ?= latest
PLATFORMS := linux
os = $(word 1, $@)

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

# default target builds all binaries for local development/testing
default: all

.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p release/$(BINARY)-$(os)-x64-v$(VERSION)/
		GOOS=$(os) GOARCH=amd64 GO111MODULE=on go build -o release/$(BINARY)-$(os)-x64-v$(VERSION)/ ./...

.PHONY: release
release: godbledger-linux-arm godbledger-darwin godbledger-windows

PHONY: clean
clean:
	rm -rf build/
	rm -rf release/
	rm -rf cert/

all:
	GO111MODULE=on go run utils/ci.go build

lint:
	GO111MODULE=on go run utils/ci.go lint

# our tests include an integration test which expects the local
# GOOS-based build output to be in the ./build/bin folder
test: all
	GO111MODULE=on go run utils/ci.go test

travis: all
	GO111MODULE=on go run utils/ci.go test -coverage $$TEST_PACKAGES

linux-arm-7:
		mkdir -p release/$(BINARY)-arm7-v$(VERSION)/
		env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on go build -o release/$(BINARY)-arm7-v$(VERSION)/ ./...

linux-arm-64:
		mkdir -p release/$(BINARY)-arm64-v$(VERSION)/
		env CC=aarch64-linux-gnu-gcc CXX=aarch-linux-gnu-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm64 GO111MODULE=on go build -o release/$(BINARY)-arm64-v$(VERSION)/ ./...

godbledger-linux-arm: godbledger-linux-arm-5 godbledger-linux-arm-6 godbledger-linux-arm-7 godbledger-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm

godbledger-linux-arm-5:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./godbledger
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm-5

godbledger-linux-arm-6:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./godbledger
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm-6

godbledger-linux-arm-7:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./godbledger
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm-7

godbledger-linux-arm64:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./godbledger
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm64

godbledger-darwin:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=darwin/* -v ./godbledger

godbledger-windows:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=windows/* -v ./godbledger

.PHONY: cert
cert:
	mkdir -p cert/
	cd cert; ../utils/gen.sh; cd ..
