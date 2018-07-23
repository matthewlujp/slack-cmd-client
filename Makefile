
.PHONY: install, build, distribute, clean, test
install: build distribute clean ;

build:
	dep ensure
	cd src && go build -o ../slack-cli .

distribute:
ifdef GOPATH
	cp slack-cli "$(GOPATH)/bin"
else
	@echo "GOPATH is not set.\nCancel installation."
endif
	
clean:
	rm slack-cli

test:
	@echo "test not defined yet"
