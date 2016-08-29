.PHONY: clobber

emojitrack-fakefeeder: main.go redis.go data.go snapshot_data.go
	go build

snapshot_data.go: scripts/generate_data.rb
	ruby $< > $@

clobber:
	rm -fv snapshot_data.go
