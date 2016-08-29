FROM scratch
MAINTAINER Matthew Rothenberg <mroth@mroth.info>

COPY build/linux-amd64/emojitrack-fakefeeder /
ENTRYPOINT ["/emojitrack-fakefeeder"]
