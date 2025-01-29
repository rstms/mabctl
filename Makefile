
bin = mabctll

$(bin): fmt
	fix go build

run:
	./$(bin)

fmt:
	fix go fmt ./...

clean:
	go clean

install: $(bin)
	go install

# install -o root -g root -m 0755 ./$(bin) /usr/local/bin/$(bin)
