BUILD_TOOL=GARBLE_EXPERIMENTAL_CONTROLFLOW=1 garble -literals -tiny -seed=AUDIRS8V12TURBO
# BUILD_TOOL=go
BUILD_FLAG=-ldflags="-w -s"
server:
	# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(BUILD_TOOL) build $(BUILD_FLAG) -o bin/server-darwin cmd/server/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_TOOL) build $(BUILD_FLAG) -o bin/server-linux cmd/server/*.go

client:
	# CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(BUILD_TOOL) build $(BUILD_FLAG) -o bin/client-darwin cmd/client/*.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_TOOL) build $(BUILD_FLAG) -o bin/core cmd/client/main.go cmd/client/hconsole.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(BUILD_TOOL) build -ldflags="-w -s -H windowsgui" -o bin/client-win.exe  cmd/client/main.go cmd/client/hconsole_windows.go
compression:
	upx --lzma -o bin/setup.exe bin/client-win.exe -f
chmod:
	# chmod +x bin/*

all: client server compression chmod