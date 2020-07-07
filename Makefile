BINARY := godbledger
VERSION ?= latest
PLATFORMS := linux
os = $(word 1, $@)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
		mkdir -p release/$(BINARY)-$(os)-x64-v$(VERSION)/
		GOOS=$(os) GOARCH=amd64 go build -o release/$(BINARY)-$(os)-x64-v$(VERSION)/ ./...

.PHONY: release
release: linux

PHONY: clean
clean:
	rm -rf release/

travis:
	GO111MODULE=on go run utils/ci.go install
	GO111MODULE=on go run utils/ci.go test -coverage $$TEST_PACKAGES
