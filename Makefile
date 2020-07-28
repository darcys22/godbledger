BINARY := godbledger
VERSION ?= latest
PLATFORMS := linux
os = $(word 1, $@)

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
		GOOS=linux GOARCH=arm GOARM=5 GO111MODULE=on go build -o release/$(BINARY)-$(os)-x64-v$(VERSION)/ ./...

