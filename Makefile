.PHONY: all test lint frontend-build build-mac build-win installer-win app-mac dmg-mac clean

APP_NAME = niuZip

all: test frontend-build build-mac app-mac

test:
	@echo "==> Running unit tests and coverage..."
	go test -v -cover ./...

lint:
	@echo "==> Checking go fmt and vet..."
	go fmt ./...
	go vet ./...

frontend-build:
	@echo "==> Building Vue 3 + Vite frontend..."
	cd frontend && npm run build

build-mac: frontend-build
	@echo "==> Building macOS Apple Silicon (arm64) CLI binary..."
	@mkdir -p build/bin
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags "production" -ldflags="-s -w -extldflags '-framework UniformTypeIdentifiers'" -o build/bin/$(APP_NAME)-darwin-arm64 .

app-mac: frontend-build
	@echo "==> Packaging macOS Apple Silicon (arm64) niuZip.app bundle..."
	@mkdir -p build/niuZip.app/Contents/MacOS build/niuZip.app/Contents/Resources
	@cp assets/macos/Info.plist build/niuZip.app/Contents/Info.plist
	@if [ -f assets/icons/appicon.icns ]; then cp assets/icons/appicon.icns build/niuZip.app/Contents/Resources/appicon.icns; fi
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags "production" -ldflags="-s -w -extldflags '-framework UniformTypeIdentifiers'" -o build/niuZip.app/Contents/MacOS/niuZip .
	codesign --force --deep -s - build/niuZip.app
	@echo "==> niuZip.app successfully packaged at build/niuZip.app"

dmg-mac: app-mac
	@echo "==> Packaging DMG installer with drag-to-Applications layout..."
	@chmod +x assets/dmg/package_dmg.sh
	@./assets/dmg/package_dmg.sh

build-win: frontend-build
	@echo "==> Cross-compiling Windows x64 binary..."
	@mkdir -p build/bin
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags "production" -ldflags="-s -w -H windowsgui" -o build/bin/$(APP_NAME)-windows-amd64.exe .

installer-win: build-win
	@echo "==> Building Windows x64 Setup installer with NSIS..."
	@which makensis >/dev/null || (echo "Error: makensis not found. Run 'brew install makensis'" && exit 1)
	cd assets/windows && makensis installer.nsi
	@echo "==> Windows Setup installer successfully created at build/$(APP_NAME)-Setup-x64.exe"

clean:
	@echo "==> Cleaning build directory..."
	rm -rf build/*
