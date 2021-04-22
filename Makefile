
EXECUTABLE := qitmeer-wallet
GITVER := $(shell git rev-parse --short=7 HEAD )
GITDIRTY := $(shell git diff --quiet || echo '-dirty')
GITVERSION = "$(GITVER)$(GITDIRTY)"
DEV=dev
RELEASE=release
LDFLAG_DEV = -X github.com/Qitmeer/qitmeer-wallet/version.AppBuild=$(DEV)-$(GITVERSION)
LDFLAG_RELEASE = -X github.com/Qitmeer/qitmeer-wallet/version.AppBuild=$(RELEASE)-$(GITVERSION)
GOFLAGS_DEV = -ldflags "$(LDFLAG_DEV)"
GOFLAGS_RELEASE = -ldflags "$(LDFLAG_RELEASE)"
VERSION=$(shell ./build/bin/qitmeer-wallet --version | grep ^qitmeer-wallet | cut -d' ' -f3|cut -d'+' -f1)
GOBIN = ./build/bin

UNIX_EXECUTABLES := \
	build/release/darwin/amd64/bin/$(EXECUTABLE) \
	build/release/linux/amd64/bin/$(EXECUTABLE)
WIN_EXECUTABLES := \
	build/release/windows/amd64/bin/$(EXECUTABLE).exe

EXECUTABLES=$(UNIX_EXECUTABLES) $(WIN_EXECUTABLES)

COMPRESSED_EXECUTABLES=$(UNIX_EXECUTABLES:%=%.tar.gz) $(WIN_EXECUTABLES:%.exe=%.zip) $(WIN_EXECUTABLES:%.exe=%.cn.zip)

RELEASE_TARGETS=$(EXECUTABLES) $(COMPRESSED_EXECUTABLES)

build:
	go build -o $(GOBIN)/qitmeer-wallet $(GOFLAGS_DEV) "github.com/Qitmeer/qitmeer-wallet"

# amd64 release
build/release/%: OS=$(word 3,$(subst /, ,$(@)))
build/release/%: ARCH=$(word 4,$(subst /, ,$(@)))
build/release/%/$(EXECUTABLE):
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer-wallet"
build/release/%/$(EXECUTABLE).exe:
	@echo Build $(@)
	@GOOS=$(OS) GOARCH=$(ARCH) go build $(GOFLAGS_RELEASE) -o $(@) "github.com/Qitmeer/qitmeer-wallet"


%.zip: %.exe
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).zip "$<"
%.cn.zip: %.exe
	@echo Build $(@).cn.zip
	@echo zip $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@zip -j $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).cn.zip "$<" script/win/start.bat

%.tar.gz : %
	@echo tar $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH)
	@tar -zcvf $(EXECUTABLE)-$(VERSION)-$(OS)-$(ARCH).tar.gz "$<"

release:cleanBuild build
	@echo "Build release version : $(VERSION)"
	@$(MAKE) $(RELEASE_TARGETS)
	@shasum -a 512 $(EXECUTABLES) > $(EXECUTABLE)-$(VERSION)_checksum.txt
	@shasum -a 512 $(EXECUTABLE)-$(VERSION)-* >> $(EXECUTABLE)-$(VERSION)_checksum.txt
checksum:
	@cat $(EXECUTABLE)-$(VERSION)_checksum.txt|shasum -c
cleanBuild:
	@rm -f *.zip
	@rm -f *.tar.gz
	@rm -f ./build/bin/qitmeer-waellt
	@rm -rf ./build/release

webui: statik npm
	cd assets/src && npm run build
	cd assets/statik && rm -fr statik.go
	cd assets && $$(go env GOPATH)/bin/statik -src ./src/dist/
statik:
	go get github.com/rakyll/statik
npm:
	if [ ! -d assets/src/node_modules/ ]; then \
		cd assets/src && npm install; \
	fi
clean:
	rm -rf assets/statik/statik.go
	rm -rf assets/src/dist
	rm -rf assets/src/node_modules/
	rm -rf assets/src/package-lock.json