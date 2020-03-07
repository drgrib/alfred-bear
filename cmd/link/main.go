package main

import (
	"fmt"
	"os"

	"golang.org/x/text/unicode/norm"

	"github.com/drgrib/alfred-bear/db"
)

func main() {
	noteID := norm.NFC.String(os.Args[1])

	litedb, err := db.NewBearDB()
	if err != nil {
		panic(err)
	}

	callback := fmt.Sprintf("bear://x-callback-url/open-note?id=%s&show_window=yes&new_window=yes", noteID)
	rows, err := litedb.Query(fmt.Sprintf(db.NOTE_TITLE_BY_ID, noteID))
	if err != nil {
		panic(err)
	}
	title := rows[0][db.TitleKey]
	link := fmt.Sprintf("[%s](%s)", title, callback)

	fmt.Print(link)
}
