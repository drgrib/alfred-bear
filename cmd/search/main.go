package main

import (
	"fmt"
	"os"
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

func addNoteRowsToAlfred(rows []map[string]string) {
	for _, row := range rows {
		alfred.Add(alfred.Item{
			Title:    row[db.TitleKey],
			Subtitle: getUniqueTagString(row[db.TagsKey]),
			Arg:      row[db.NoteIDKey],
			Valid:    alfred.Bool(true),
		})
	}
}

func main() {
	query := os.Args[1]

	litedb, err := db.NewBearDB()
	if err != nil {
		panic(err)
	}

	elements := strings.Split(query, " ")
	tags := []string{}
	words := []string{}
	lastElement := ""
	for _, e := range elements {
		switch {
		case e == "":
		case strings.HasPrefix(e, "#"):
			tags = append(tags, e)
		default:
			words = append(words, e)
		}
		lastElement = e
	}

	wordStr := strings.Join(words, " ")

	switch {
	case strings.HasPrefix(lastElement, "#"):
		rows, err := litedb.Query(fmt.Sprintf(db.TAGS_BY_TITLE, lastElement[1:]))
		if err != nil {
			panic(err)
		}

		for _, row := range rows {
			tag := "#" + row[db.TitleKey]
			autocomplete := strings.Join(elements[:len(elements)-1], " ") + " " + tag + " "
			alfred.Add(alfred.Item{
				Title:        tag,
				Autocomplete: autocomplete,
				Valid:        alfred.Bool(false),
			})
		}

	case len(words) == 0 && len(tags) == 0 && lastElement == "":
		rows, err := litedb.Query(db.RECENT_NOTES)
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)

	case len(tags) != 0:
		tagConditions := []string{}
		for _, t := range tags {
			c := fmt.Sprintf("lower(tag.ZTITLE) = lower('%s')", t[1:])
			tagConditions = append(tagConditions, c)
		}
		tagConjunction := strings.Join(tagConditions, " OR ")
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_TAGS_AND_QUERY, tagConjunction, wordStr, wordStr, len(tags), wordStr))
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)

	default:

		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_QUERY, wordStr, wordStr, wordStr))
		if err != nil {
			panic(err)
		}
		addNoteRowsToAlfred(rows)
	}

	alfred.Run()
}
