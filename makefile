GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install

DISTRIBUTION_DIR=dist
PLATFORMS := linux/amd64 windows/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

exlinux=
exwindows=.exe

vendoring:
	$(GOGET) github.com/kardianos/govendor
	govendor sync

$(PLATFORMS):
	$(info Building for $(os))
	GOOS=$(os) GOARCH=$(arch) go build -o 'dist/$(os)/$(arch)/sauron$(ex$(os))' sauron.go

install:
	$(GOCMD) install
justbuildit: $(PLATFORMS)
release: vendoring $(PLATFORMS)
all: vendoring install