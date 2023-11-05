package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/drgrib/alfred-bear/db"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	litedb, err := db.NewBearDB()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	targetTags := []string{"#health/sleep", "#media/youtube"}
	notes, err := litedb.QueryNotesByTextAndTags("", targetTags)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	for _, n := range notes {
		tags := strings.Split(n[db.TagsKey], ",")
		found := map[string]bool{}
		for _, t := range tags {
			found[t] = true
		}
		invalid := false
		for _, t := range targetTags {
			if !found[db.RemoveTagHashes(t)] {
				invalid = true
			}
		}
		if invalid {
			fmt.Print("INVALID ")
		}
		fmt.Println(n[db.TitleKey])
	}
}
