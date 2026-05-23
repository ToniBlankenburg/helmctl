BINARY  := helmctl
PREFIX  ?= $(HOME)/.local/bin

.PHONY: build install uninstall

build:
	go build -o $(BINARY) .

install: build
	mkdir -p $(PREFIX)
	install -m 755 $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	rm -f $(PREFIX)/$(BINARY)
