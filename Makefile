
sourcePath=./cmd/services-match/main.go

build-services-match:
		go build -trimpath -o bin/services-match ${sourcePath}
build-services-match-linux:
    	GOOS=linux GOARCH=amd64 go build -trimpath -o bin/services-match-linux ${sourcePath}
build-services-match-win:
    	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -trimpath -o bin/services-match-win ${sourcePath}