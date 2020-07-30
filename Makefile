BINARY := godbledger
VERSION ?= latest
PLATFORMS := linux
os = $(word 1, $@)

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p release/$(BINARY)-$(os)-x64-v$(VERSION)/
		GOOS=$(os) GOARCH=amd64 GO111MODULE=on go build -o release/$(BINARY)-$(os)-x64-v$(VERSION)/ ./...

.PHONY: release
release: linux

PHONY: clean
clean:
	rm -rf release/

travis:
	GO111MODULE=on go run utils/ci.go install
	GO111MODULE=on go run utils/ci.go test -coverage $$TEST_PACKAGES

arm:
		mkdir -p release/$(BINARY)-arm-x64-v$(VERSION)/
		env CC=arm-linux-gnueabihf-gcc CXX=arm-linux-gnueabihf-g++ CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 GO111MODULE=on go build -o release/$(BINARY)-$(os)-x64-v$(VERSION)/ ./...

godbledger-linux-arm: godbledger-linux-arm-5 godbledger-linux-arm-6 godbledger-linux-arm-7 godbledger-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/godbledger-linux-* | grep arm

godbledger-linux-arm-5:
	$(GORUN) utils/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./..
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
