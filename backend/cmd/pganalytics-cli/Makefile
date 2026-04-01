.PHONY: build install clean test

VERSION := 0.1.0
BINARY_NAME := pganalytics

build:
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) main.go

install: build
	cp $(BINARY_NAME) ${HOME}/.local/bin/$(BINARY_NAME)
	chmod +x ${HOME}/.local/bin/$(BINARY_NAME)
	@echo "✓ Installed to ~/.local/bin/$(BINARY_NAME)"

test:
	go test ./tests -v

clean:
	rm -f $(BINARY_NAME)
	go clean
