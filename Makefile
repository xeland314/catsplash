BIN_DIR=bin
BINARY=$(BIN_DIR)/catsplash
CTL_BINARY=$(BIN_DIR)/catsctl
GOFILES=main.go

build:
	mkdir -p $(BIN_DIR)
	rm -f $(BINARY) $(CTL_BINARY) catsctl/main
	CGO_ENABLED=1 go build -o $(BINARY) $(GOFILES)
	CGO_ENABLED=1 go build -o $(CTL_BINARY) catsctl/main.go

test:
	go test -v ./...

clean:
	rm -f $(BINARY) $(CTL_BINARY) catsctl/main *.db

run: build
	sudo ./$(BINARY)

.PHONY: build test clean run
