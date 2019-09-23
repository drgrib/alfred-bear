package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/atotto/clipboard"
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

	if !autocompleted {
		callback := []string{}
		if q.WordString != "" {
			callback = append(callback, "title="+url.PathEscape(q.WordString))
		}
		if len(q.Tags) != 0 {
			bareTags := []string{}
			for _, t := range q.Tags {
				bareTags = append(bareTags, url.PathEscape(t[1:]))
			}
			callback = append(callback, "tags="+strings.Join(bareTags, ","))
		}

		clipString, err := clipboard.ReadAll()
		if err != nil {
			panic(err)
		}
		if clipString != "" {
			callback = append(callback, "&text="+url.PathEscape(clipString))
		}
		callbackString := strings.Join(callback, "&")

		item := alfred.Item{
			Title: fmt.Sprintf("Create %#v", q.WordString),
			Arg:   callbackString,
			Valid: alfred.Bool(true),
		}
		if len(q.Tags) != 0 {
			item.Subtitle = strings.Join(q.Tags, " ")
		}
		alfred.Add(item)
	}

	alfred.Run()
}
