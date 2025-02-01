
bin = mabctll
version != cat VERSION
latest_release := $(shell gh release view --json tagName --jq .tagName)

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

.release: $(bin)
ifeq "v$(version)" "$(latest_release)"
	@echo version $(version) is already released
else
	gh release create v$(version) --generate-notes --target master;
endif
	@touch $@

latest-release:
	@echo $(latest_release)

release: .release

