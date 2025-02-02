
bin = mabctll

$(bin): fmt
	fix go build

run:
	./$(bin)

fix:
	fix go fmt ./...

fmt:
	go fmt ./...

clean:
	go clean

install: $(bin)
	go install

release:
	gh release create v$(shell cat VERSION) --generate-notes --target master
