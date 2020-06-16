.PHONY: image clean clobber

app        := fakefeeder
docker-tag := emojitracker/fakefeeder
cmd        := ./cmd/fakefeeder

src := $(cmd)/main.go $(cmd)/redis.go data.go feeder.go rankings/rankings.go rankings/snapshot.go

default: bin/$(app)

bin/$(app): $(src)
	go build -o $@ $(cmd)

image: $(src) Dockerfile
	DOCKER_BUILDKIT=1 docker build -t $(docker-tag) .

rankings/snapshot.go: rankings/scripts/generate_snapshot.go
	go generate ./rankings

clean:
	rm -rf bin

clobber: clean
	rm -f rankings/snapshot.go
