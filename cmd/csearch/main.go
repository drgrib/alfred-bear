package main

import (
	"os"

	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/core"
	"github.com/drgrib/alfred-bear/db"
)

func main() {
	createIndex := 2
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
		searchRows, err := core.GetSearchRows(litedb, query)
		if err != nil {
			panic(err)
		}
		createItem, err := core.GetCreateItem(query)
		if err != nil {
			panic(err)
		}

		for _, row := range searchRows[:createIndex] {
			alfred.Add(core.RowToItem(row))
		}
		alfred.Add(*createItem)
		for _, row := range searchRows[createIndex:] {
			alfred.Add(core.RowToItem(row))
		}
	}

	alfred.Run()
}
