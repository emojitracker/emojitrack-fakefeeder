.PHONY: image clean clobber

app        := fakefeeder
docker-tag := emojitracker/fakefeeder
cmd        := ./cmd/fakefeeder

src := $(cmd)/main.go $(cmd)/redis.go data.go feeder.go snapshot_data.go

default: bin/$(app)

bin/$(app): $(src)
	go build -o $@ $(cmd)

image: $(src) Dockerfile
	DOCKER_BUILDKIT=1 docker build -t $(docker-tag) .

snapshot_data.go: scripts/generate_snapshot.go
	go generate

clean:
	rm -rf bin

clobber: clean
	rm -f snapshot_data.go
