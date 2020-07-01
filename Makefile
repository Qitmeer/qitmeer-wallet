build: webui
	go build

webui: statik npm
	cd assets/src && npm run build
	cd assets/statik && rm -fr statik.go
	cd assets && $$(go env GOPATH)/bin/statik -src ./src/dist/
statik:
	go get github.com/rakyll/statik
npm:
	if [ ! -d assets/src/node_modules/ ]; then \
		cd assets/src && npm install; \
	fi
clean:
	rm -rf assets/statik/statik.go
	rm -rf assets/src/dist
	rm -rf assets/src/node_modules/
	rm -rf assets/src/package-lock.json
