.PHONY: image clean clobber

app        := emojitrack-fakefeeder
docker-tag := emojitracker/fakefeeder

src := main.go redis.go data.go feeder.go snapshot_data.go

default: bin/$(app)

bin/$(app): $(src)
	go build -o $@

image: $(src) Dockerfile
	DOCKER_BUILDKIT=1 docker build -t $(docker-tag) .

snapshot_data.go: scripts/generate_snapshot.go
	go generate

clean:
	rm -rf bin

clobber: clean
	rm -f snapshot_data.go
