package core

import (
	"fmt"
	"strings"

	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/db"
)

func getUniqueTagString(tagString string) string {
	tags := strings.Split(tagString, ",")
	uniqueTags := []string{}
	for _, t := range tags {
		isPrefix := false
		for _, other := range tags {
			if t != other && strings.HasPrefix(other, t) {
				isPrefix = true
				break
			}
		}
		if !isPrefix {
			uniqueTags = append(uniqueTags, t)
		}
	}
	return "#" + strings.Join(uniqueTags, " #")
}

func AddNoteRowsToAlfred(rows []map[string]string) {
	for _, row := range rows {
		alfred.Add(alfred.Item{
			Title:    row[db.TitleKey],
			Subtitle: getUniqueTagString(row[db.TagsKey]),
			Arg:      row[db.NoteIDKey],
			Valid:    alfred.Bool(true),
		})
	}
}

type Query struct {
	Tokens, Tags          []string
	LastToken, WordString string
}

func ParseQuery(query string) Query {
	q := Query{}
	q.Tokens = strings.Split(query, " ")
	q.Tags = []string{}
	words := []string{}
	for _, e := range q.Tokens {
		switch {
		case e == "":
		case strings.HasPrefix(e, "#"):
			q.Tags = append(q.Tags, e)
		default:
			words = append(words, e)
		}
	}
	q.LastToken = q.Tokens[len(q.Tokens)-1]
	q.WordString = strings.Join(words, " ")
	return q
}

func AutocompleteTags(litedb db.LiteDB, q Query) (bool, error) {
	if strings.HasPrefix(q.LastToken, "#") {
		rows, err := litedb.Query(fmt.Sprintf(db.TAGS_BY_TITLE, q.LastToken[1:]))
		if err != nil {
			return false, err
		}

		for _, row := range rows {
			tag := "#" + row[db.TitleKey]
			autocomplete := strings.Join(q.Tokens[:len(q.Tokens)-1], " ") + " " + tag + " "
			alfred.Add(alfred.Item{
				Title:        tag,
				Autocomplete: autocomplete,
				Valid:        alfred.Bool(false),
			})
		}
		return true, nil
	}
	return false, nil
}
