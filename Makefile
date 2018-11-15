HAS_GLIDE := $(shell command -v glide;)
VERSION := 1.0
DIST := $(CURDIR)/_dist
BUILD := $(CURDIR)/_build
LDFLAGS := "-X main.version=${VERSION}"
BINARY := m2-helper

.PHONY: build
build: bootstrap
	go build -o $(BUILD)/$(BINARY) -ldflags $(LDFLAGS)

# build static binaries: https://medium.com/@diogok/on-golang-static-binaries-cross-compiling-and-plugins-1aed33499671
.PHONY: dist
dist: bootstrap
	mkdir -p $(BUILD)
	mkdir -p $(DIST)
	cp README.md $(BUILD) && cp LICENSE $(BUILD)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD)/$(BINARY) -ldflags $(LDFLAGS) -a -tags netgo
	tar -C $(BUILD) -zcvf $(DIST)/helm-$(BINARY)-linux-$(VERSION).tgz $(BINARY) README.md LICENSE
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $(BUILD)/$(BINARY) -ldflags $(LDFLAGS) -a -tags netgo
	tar -C $(BUILD) -zcvf $(DIST)/helm-$(BINARY)-macos-$(VERSION).tgz $(BINARY) README.md LICENSE
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $(BUILD)/$(BINARY).exe -ldflags $(LDFLAGS) -a -tags netgo
	tar -C $(BUILD) -llzcvf $(DIST)/helm-$(BINARY)-windows-$(VERSION).tgz $(BINARY).exe README.md LICENSE

.PHONY: bootstrap
bootstrap:
ifndef HAS_GLIDE
	go get -u github.com/Masterminds/glide
endif
	glide install --strip-vendor

.PHONY: clean
clean:
	rm -rf _*