
bin = mabctll

$(bin): go.sum 
	fix go build

run:
	./$(bin)

fix:
	fix go fmt ./...

fmt:
	go fmt ./...

clean:
	go clean

test:
	go test -v -failfast . ./...

sterile: clean
	go clean -cache -modcache -testcache -i
	rm -f go.mod
	rm -f go.sum


go.sum: go.mod
	go mod tidy

go.mod:
	go mod init

install: $(bin)
	go install

release:
	gh release create v$(shell cat VERSION) --generate-notes --target master

update:
	go get github.com/rstms/go-webdav@$(shell gh --repo rstms/go-webdav release list | awk '{print $$1;exit}')
