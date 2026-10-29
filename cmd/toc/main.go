package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
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

	rows, err := litedb.Query(fmt.Sprintf(db.NOTE_TEXT_BY_ID, noteID))
	if err != nil {
		panic(err)
	}
	text := rows[0][db.TextKey]

	headerRe := regexp.MustCompile(`^#+ .+`)
	var tocLines []string
	textLines := strings.Split(text, "\n")
	for _, l := range textLines {
		if headerRe.MatchString(l) {
			hashCount := 0
			for _, r := range l {
				if r != '#' {
					break
				}
				hashCount++
			}

			header := strings.TrimLeft(l, "# ")
			escaped := url.PathEscape(header)
			callback := fmt.Sprintf(
				"bear://x-callback-url/open-note?id=%s&header=%s&show_window=yes&new_window=yes", noteID, escaped)

			header = strings.ReplaceAll(header, "[", "- ")
			header = strings.ReplaceAll(header, "]", " -")

			tocLines = append(tocLines,
				fmt.Sprintf("%s* [%s](%s)", strings.Repeat("\t", hashCount-1), header, callback),
			)
		}
	}

	toc := strings.Join(tocLines, "\n")
	fmt.Print(toc)
}
