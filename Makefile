.PHONY: container clean clobber

app        := emojitrack-fakefeeder
static-app := build/linux-amd64/$(app)
docker-tag := emojitracker/fakefeeder

src := main.go redis.go data.go snapshot_data.go

bin/$(app): $(src)
	go build -o $@

$(static-app): $(src)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags "-s" -a -installsuffix cgo -o $(static-app)

container: $(static-app)
	docker build -t $(docker-tag) .

snapshot_data.go: scripts/generate_data.rb
	ruby $< > $@

clean:
	rm -rf bin build

clobber: clean
	rm -f snapshot_data.go
