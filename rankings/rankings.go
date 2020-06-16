package rankings

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	fakefeeder "github.com/emojitracker/emojitrack-fakefeeder"
)

// public API endpoints
const (
	EmojitrackerV1APIRankingsURL = "https://api.emojitracker.com/v1/rankings"
)

// Live retrieves the live rankings from the provided Emojitracker API
// endpoint url.
//
// This is primarily used when generating a new snapshot, but is provided
// in the public methods in case fresher data is needed for some reason.
func Live(url string) (results []fakefeeder.Ranking, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = errors.New("not OK status when retrieving remote rankings")
		return
	}

	defer resp.Body.Close()
	dat, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(dat, &results)
	return
}

//go:generate go run ./scripts/generate_snapshot.go -- snapshot.go
