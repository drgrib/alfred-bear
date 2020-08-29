package main

import (
	"log"
	"os"
	"strings"

	"github.com/drgrib/alfred"

	"github.com/drgrib/alfred-bear/core"
	"github.com/drgrib/alfred-bear/db"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	createIndex := 2
	query := core.ParseQuery(os.Args[1])

	litedb, err := db.NewBearDB()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	autocompleted, err := core.Autocomplete(litedb, query)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	if !autocompleted {
		searchRows, err := core.GetSearchRows(litedb, query)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		appSearchItem, err := core.GetAppSearchItem(query)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		createItem, err := core.GetCreateItem(query)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		if len(searchRows) > 0 {
			endIndex := createIndex
			if len(searchRows) < createIndex {
				endIndex = len(searchRows)
			}
			for _, row := range searchRows[:endIndex] {
				alfred.Add(core.RowToItem(row, query))
			}
		} else if strings.Contains(query.WordString, "@") {
			alfred.Add(*appSearchItem)
		}
		alfred.Add(*createItem)
		if len(searchRows) > createIndex {
			for _, row := range searchRows[createIndex:] {
				alfred.Add(core.RowToItem(row, query))
			}
		}
	}

	alfred.Run()
}
