COVER_PROFILE=cover.out
COVER_HTML=cover.html

.PHONY: $(COVER_PROFILE) $(COVER_HTML)

all: package

build: clean
	mkdir -p -v ./bin/amm.app/Contents/Resources/assets/icon
	mkdir -p -v ./bin/amm.app/Contents/MacOS
	cp ./appInfo/*.plist ./bin/amm.app/Contents/Info.plist
	cp ./appInfo/*.icns ./bin/amm.app/Contents/Resources/icon.icns
	cp ./assets/icon/* ./bin/amm.app/Contents/Resources/assets/icon
	go build -o ./bin/amm.app/Contents/MacOS/amm cmd/main.go

package: build
	ln -s /Applications ./bin/Applications
	hdiutil create -volname "Automatic Mouse Mover" -srcfolder ./bin -ov -format UDZO AutomaticMouseMover.dmg

clean:
	rm -rf ./bin

start:
	go run cmd/main.go

test:coverage

coverage: $(COVER_HTML)

$(COVER_HTML): $(COVER_PROFILE)
	go tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

$(COVER_PROFILE):
	go test -v -failfast -race -coverprofile=$(COVER_PROFILE) ./...

vet:
	go vet $(shell glide nv)

lint:
	go list ./... | grep -v vendor | grep -v /assets/ |xargs -L1 golint -set_exit_status

.PHONY: build 
.PHONY: clean