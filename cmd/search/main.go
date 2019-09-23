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

	q := core.ParseQuery(query)

	autocompleted, err := core.AutocompleteTags(litedb, q)
	if err != nil {
		panic(err)
	}

	switch {
	case autocompleted:
		// short-circuit others
	case q.WordString == "" && len(q.Tags) == 0 && q.LastToken == "":
		rows, err := litedb.Query(db.RECENT_NOTES)
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)

	case len(q.Tags) != 0:
		tagConditions := []string{}
		for _, t := range q.Tags {
			c := fmt.Sprintf("lower(tag.ZTITLE) = lower('%s')", t[1:])
			tagConditions = append(tagConditions, c)
		}
		tagConjunction := strings.Join(tagConditions, " OR ")
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_TAGS_AND_QUERY, tagConjunction, q.WordString, q.WordString, len(q.Tags), q.WordString))
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)

	default:
		rows, err := litedb.Query(fmt.Sprintf(db.NOTES_BY_QUERY, q.WordString, q.WordString, q.WordString))
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)
	}

	alfred.Run()
}
