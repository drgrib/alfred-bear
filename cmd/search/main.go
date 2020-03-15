package main

import (
	"os"

	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/core"
	"github.com/drgrib/alfred-bear/db"
)

func main() {
	query := core.ParseQuery(os.Args[1])

	litedb, err := db.NewBearDB()
	if err != nil {
		panic(err)
	}

	autocompleted, err := core.AutocompleteTags(litedb, query)
	if err != nil {
		panic(err)
	}

	if !autocompleted {
		rows, err := core.GetSearchRows(litedb, query)
		if err != nil {
			panic(err)
		}
		core.AddNoteRowsToAlfred(rows)
		if len(rows) == 0 {
			alfred.Add(alfred.Item{
				Title: "No matching items found",
				Valid: alfred.Bool(false),
			})
		}
	}

	alfred.Run()
}
