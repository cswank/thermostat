pi:
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -tags netgo
macos:
	GOOS=darwin GOARCH=amd64 go build
