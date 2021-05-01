# emojitrack-fakefeeder :dizzy:
> Feeds passable random data to a Redis instance to emulate emojitrack-feeder

Feeds passable random data to a Redis instance to emulate [emojitrack-feeder],
so that one can do realistic local development on Emojitracker realtime APIs
without running a local feeder instance (which requires elevated access from
Twitter), or even without a network connection entirely!

Generally this is designed to be paired with a redis docker container within a
docker-compose workflow, e.g. when you are emulating the topology of the
Emojitracker backend infrastructure within a virtual Docker Network, this
can be swapped in transparently for the actual `emojitrack-feeder` container.

[emojitrack-feeder]: https://github.com/mroth/emojitrack-feeder

[![Docker Build Status](https://img.shields.io/docker/cloud/build/emojitracker/fakefeeder.svg?style=flat-square)](https://hub.docker.com/r/emojitracker/fakefeeder/)

## Usage

    Usage of emojitrack-fakefeeder:
      -rate int
          number of updates per second to generate (default 250)
      -target string
          URI for redis target (default "redis://localhost:6379")
      -v	verbose log all updates to stdout
      -weight
          weight random emoji probability based on history (default true)
