BINARY=catsplash
CTL_BINARY=catsctl
GOFILES=main.go

build:
	CGO_ENABLED=1 go build -o $(BINARY) $(GOFILES)
	CGO_ENABLED=1 go build -o $(CTL_BINARY) catsctl/main.go

test:
	go test -v ./...

clean:
	rm -f $(BINARY) $(CTL_BINARY) *.db

run: build
	sudo ./$(BINARY)

.PHONY: build test clean run
