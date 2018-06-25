.PHONY: container clean clobber

app        := emojitrack-fakefeeder
docker-tag := emojitracker/fakefeeder

src := main.go redis.go data.go snapshot_data.go

default: bin/$(app)

vendor:
	glide install

bin/$(app): $(src) vendor
	go build -o $@

container: $(src) Dockerfile
	docker build -t $(docker-tag) .

snapshot_data.go: scripts/generate_data.rb
	ruby $< > $@

clean:
	rm -rf bin

clobber: clean
	rm -f snapshot_data.go
