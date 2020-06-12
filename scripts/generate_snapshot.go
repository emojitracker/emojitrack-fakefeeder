// build !ignore

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

const (
	rankingsURL = "https://api.emojitracker.com/v1/rankings"
)

type ranking struct {
	Char  string `json:"char"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type rankings []ranking

// Override output format for code generation, creates identical output for the
// data structure to the old hand-written Go from the previous Ruby script.
func (r rankings) GoString() string {
	var b strings.Builder
	b.WriteString("[]emojiRanking{")
	for _, v := range r {
		b.WriteString(
			fmt.Sprintf(
				"\n\t{char: \"%s\", id: \"%s\", name: \"%s\", score: %d},",
				v.Char, v.ID, v.Name, v.Score,
			),
		)
	}
	b.WriteString("\n}")
	return b.String()
}

func getRankings(url string) (results rankings, err error) {
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

const snapshotTmpl = `// Code generated via scripts/generate_snapshot.go -- DO NOT EDIT.
// Data obtained from {{.Source}} at {{ .Time.Format "2006-01-02 15:04:05 -0700" }}.

package main

var emojiRankings = {{ printf "%#v" .Data }}
`

var tmpl = template.Must(template.New("snapshot").Parse(snapshotTmpl))

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("args", flag.Args())
		fmt.Fprintln(os.Stderr, "usage:", os.Args[0], "<output-file>")
		os.Exit(1)
	}

	ranks, err := getRankings(rankingsURL)
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
		Data   rankings
	}{
		rankingsURL,
		time.Now(),
		ranks,
	}
	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatal(err)
	}
}
