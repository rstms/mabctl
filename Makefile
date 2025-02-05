
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

sterile: clean
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
