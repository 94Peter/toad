VERSION=`git describe --tags`
BUILD_TIME=`date +%FT%T%z`
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

run: clear
	go build ${LDFLAGS} -o ./bin/api ./main.go
	./bin/api 

gcp:
	gcloud app deploy

clear:
	rm -rf ./bin/$(SERVICE)