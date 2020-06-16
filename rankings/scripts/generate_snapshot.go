// build !ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	fakefeeder "github.com/emojitracker/emojitrack-fakefeeder"
	rankings "github.com/emojitracker/emojitrack-fakefeeder/rankings"
)

type rankingCollection []fakefeeder.Ranking

// Override output format for code generation, creates identical output for the
// data structure to the old hand-written Go from the previous Ruby script.
func (r rankingCollection) GoString() string {
	var b strings.Builder
	b.WriteString("[]fakefeeder.Ranking{")
	for _, v := range r {
		b.WriteString(
			fmt.Sprintf(
				"\n\t{Char: \"%s\", ID: \"%s\", Name: \"%s\", Score: %d},",
				v.Char, v.ID, v.Name, v.Score,
			),
		)
	}
	b.WriteString("\n}")
	return b.String()
}

const snapshotTmpl = `// Code generated via scripts/generate_snapshot.go -- DO NOT EDIT.
// Data obtained from {{.Source}} at {{ .Time.Format "2006-01-02 15:04:05 -0700" }}.

package rankings

import "github.com/emojitracker/emojitrack-fakefeeder"

// Snapshot returns the archived snapshot of rankings from the Emojitracker API.
func Snapshot() []fakefeeder.Ranking {
	return snapshotData
}

var snapshotData = {{ printf "%#v" .Data }}
`

var tmpl = template.Must(template.New("snapshot").Parse(snapshotTmpl))

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("args", flag.Args())
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "<output-file>")
		os.Exit(1)
	}

	ranks, err := rankings.Live(rankings.EmojitrackerV1APIRankingsURL)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	data := struct {
		Source string
		Time   time.Time
		Data   rankingCollection
	}{
		rankings.EmojitrackerV1APIRankingsURL,
		time.Now(),
		ranks,
	}
	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatal(err)
	}
}
