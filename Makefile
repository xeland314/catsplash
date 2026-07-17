BIN_DIR=bin
BINARY=$(BIN_DIR)/catsplash
CTL_BINARY=$(BIN_DIR)/catsctl
GOFILES=main.go
LDFLAGS=-s -w

build:
	mkdir -p $(BIN_DIR)
	rm -f $(BINARY) $(CTL_BINARY) catsctl/main
	CGO_ENABLED=1 go build -o $(BINARY) $(GOFILES)
	CGO_ENABLED=1 go build -o $(CTL_BINARY) catsctl/main.go

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR) *.db

# Security and compliance checks
gosec:
	gosec -exclude-generated ./...

lopdp-check:
	bash .github/scripts/lopdp-check.sh

security: gosec lopdp-check

run: build
	sudo ./$(BINARY)

# Compilación cruzada y empaquetado automático para el lanzamiento v0.1.0
release: clean
	mkdir -p $(BIN_DIR)
	
	mkdir -p $(BIN_DIR)/amd64-v1
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/amd64-v1/catsplash $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/amd64-v1/catsctl catsctl/main.go
	tar -czf $(BIN_DIR)/catsplash-linux-amd64-v1.tar.gz -C $(BIN_DIR)/amd64-v1 catsplash catsctl
	rm -rf $(BIN_DIR)/amd64-v1

	mkdir -p $(BIN_DIR)/amd64-v2
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/amd64-v2/catsplash $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v2 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/amd64-v2/catsctl catsctl/main.go
	tar -czf $(BIN_DIR)/catsplash-linux-amd64-v2.tar.gz -C $(BIN_DIR)/amd64-v2 catsplash catsctl
	rm -rf $(BIN_DIR)/amd64-v2

	mkdir -p $(BIN_DIR)/386
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/386/catsplash $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/386/catsctl catsctl/main.go
	tar -czf $(BIN_DIR)/catsplash-linux-386.tar.gz -C $(BIN_DIR)/386 catsplash catsctl
	rm -rf $(BIN_DIR)/386

	mkdir -p $(BIN_DIR)/arm64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/arm64/catsplash $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/arm64/catsctl catsctl/main.go
	tar -czf $(BIN_DIR)/catsplash-linux-arm64.tar.gz -C $(BIN_DIR)/arm64 catsplash catsctl
	rm -rf $(BIN_DIR)/arm64

	mkdir -p $(BIN_DIR)/armv7
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/armv7/catsplash $(GOFILES)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="$(LDFLAGS)" -o $(BIN_DIR)/armv7/catsctl catsctl/main.go
	tar -czf $(BIN_DIR)/catsplash-linux-armv7.tar.gz -C $(BIN_DIR)/armv7 catsplash catsctl
	rm -rf $(BIN_DIR)/armv7

.PHONY: build test clean run release
