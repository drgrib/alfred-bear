package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/drgrib/alfred-bear/db"
)

func main() {
	noteID := norm.NFC.String(os.Args[1])

	litedb, err := db.NewBearDB()
	if err != nil {
		panic(err)
	}

	rows, err := litedb.Query(fmt.Sprintf(db.NOTE_TITLE_BY_ID, noteID))
	if err != nil {
		panic(err)
	}
	title := rows[0][db.TitleKey]
	title = strings.ReplaceAll(title, "[", `\[`)
	title = strings.ReplaceAll(title, "]", `\]`)
	title = strings.ReplaceAll(title, "/", `\/`)
	link := fmt.Sprintf("[[%s]]", title)

	fmt.Print(link)
}
