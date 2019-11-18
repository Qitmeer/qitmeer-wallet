build: webui
	go build

webui:
	cd assets/src && npm install 
	cd assets/src && npm run build
	cd assets/statik && rm -fr statik.go
	cd assets && statik -src ./src/dist/