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

func AutocompleteTags(litedb db.LiteDB, tokens []string) (bool, error) {
	lastToken := tokens[len(tokens)-1]
	if strings.HasPrefix(lastToken, "#") {
		rows, err := litedb.Query(fmt.Sprintf(db.TAGS_BY_TITLE, lastToken[1:]))
		if err != nil {
			return false, err
		}

		for _, row := range rows {
			tag := "#" + row[db.TitleKey]
			autocomplete := strings.Join(tokens[:len(tokens)-1], " ") + " " + tag + " "
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
