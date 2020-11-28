BINARY := godbledger
VERSION ?= latest
PLATFORMS := linux
os = $(word 1, $@)

GOBIN = ./build/bin GO ?= latest
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
	rm -rf cert/

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/transaction.proto

lint:
	GO111MODULE=on go run utils/ci.go lint

travis:
	GO111MODULE=on go run utils/ci.go install
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

.PHONY: cert
cert:
	mkdir -p cert/
	cd cert; ../utils/gen.sh; cd ..
