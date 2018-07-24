
ifndef TARGETOS
	TARGETOS = darwin
endif

ifndef TARGETARCH
	TARGETARCH = amd64

endif

ifndef OUT
	BINOUT = ../slack-cli
else
	BINOUT = ../$(OUT)
endif

.PHONY: install, build, distribute, clean, test
install: build distribute clean ;

build:
	dep ensure
	cd src && GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) go build -o $(BINOUT) .

distribute:
ifdef GOPATH
	cp slack-cli "$(GOPATH)/bin"
else
	@echo "GOPATH is not set.\nCancel installation."
endif
	
clean:
	rm slack-cli

test:
	@echo "test package slack"
	cd src/slack && go test
	@echo "\n"
	@echo "test package main"
	cd src && go test
	@echo "\n"
