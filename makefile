GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

DISTRIBUTION_DIR=dist
PLATFORMS := linux/amd64 windows/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

exlinux=
exwindows=.exe

$(PLATFORMS):
	$(info Building for $(os))
	GOOS=$(os) GOARCH=$(arch) go build -o 'dist/$(os)/$(arch)/sauron-cli$(ex$(os))' cmd/sauron-cli.go

release: $(PLATFORMS)
