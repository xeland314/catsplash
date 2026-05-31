BINARY=catsplash
GOFILES=main.go

build:
	CGO_ENABLED=1 go build -o $(BINARY) $(GOFILES)

test:
	go test -v ./...

clean:
	rm -f $(BINARY) *.db

run: build
	sudo ./$(BINARY)

.PHONY: build test clean run
