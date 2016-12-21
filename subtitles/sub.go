package subtitles

import (
	"io"
	"sort"
	"time"

	"github.com/n00bDooD/gosrt"
)

func init() {
	gosrt.InputValidationStrictness = gosrt.SkipInvalid
}

type Subtitle struct {
	Start time.Duration
	End   time.Duration
	Text  string // text to display while subtitle is on screen
}

type Subtitles []Subtitle

func (s Subtitles) Len() int           { return len(s) }
func (s Subtitles) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Subtitles) Less(i, j int) bool { return s[i].Start < s[j].Start }

func GetSubtitles(texts []Subtitle, offset time.Duration, length time.Duration) []Subtitle {
	sort.Sort(Subtitles(texts))
	var response []Subtitle
	for _, sub := range texts {
		if sub.Start < offset {
			continue
		}
		if sub.Start > offset+length {
			break
		}
		response = append(response, sub)
	}
	return response
}

func Parse(r io.Reader) ([]Subtitle, error) {
	var subs []Subtitle
	scanner := gosrt.NewScanner(r)
	for scanner.Scan() {
		subs = append(subs, from3rdParty(scanner.Subtitle()))
	}
	return subs, scanner.Err()
}

func from3rdParty(sub gosrt.Subtitle) Subtitle {
	return Subtitle{
		Start: sub.Start,
		End:   sub.End,
		Text:  sub.Text,
	}
}
