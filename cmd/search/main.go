package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/core"
	"github.com/drgrib/alfred-bear/db"
)

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

	autocompleted, err := core.AutocompleteTags(litedb, elements)
	if err != nil {
		panic(err)
	}

	switch {
	case autocompleted:
		// short-circuit others
	case wordStr == "" && len(tags) == 0 && lastElement == "":
		rows, err := litedb.Query(db.RECENT_NOTES)
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)

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
		core.AddNoteRowsToAlfred(rows)

	default:
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_QUERY, wordStr, wordStr, wordStr))
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)
	}

	alfred.Run()
}
