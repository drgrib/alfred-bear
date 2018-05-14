package db

import (
	. "fmt"
	"testing"

	"github.com/drgrib/alfred-bear/comp"
)

func TestStringQuery(t *testing.T) {
	var q = `    
	SELECT DISTINCT
		ZUNIQUEIDENTIFIER, ZTITLE 
	FROM 
		ZSFNOTE 
	WHERE 
		ZARCHIVED=0 
		AND ZTRASHED=0 
	ORDER BY 
		ZMODIFICATIONDATE DESC 
	LIMIT 25
`
	path := comp.Expanduser("~/Library/Containers/net.shinyfrog.bear/Data/Documents/Application Data/database.sqlite")
	lite, err := NewLiteDB(path)
	comp.MustBeNil(err)
	results, err := lite.QueryStringMaps(q)
	comp.MustBeNil(err)
	for _, m := range results {
		for k, v := range m {
			Println(k, v)
		}
		Println()
	}
}

func TestNoteList(t *testing.T) {
	notes := NewNoteList()
	notes.AppendNew(Note{"XXX", "Note 1"})
	notes.AppendNew(Note{"XXX", "Note 1"})
	notes.AppendNew(Note{"XX2", "Note 2"})
	notes2 := NewNoteList()
	notes2.AppendNew(Note{"XXX", "Note 1"})
	notes2.AppendNew(Note{"XX3", "Note 3"})
	notes.AppendNewFrom(notes2)
	Println(notes)
}

func TestBearDB(t *testing.T) {
	db, err := NewBearDB()
	comp.MustBeNil(err)
	tags, err := db.SearchTags("")
	comp.MustBeNil(err)
	Println(tags)
	tags, err = db.SearchTags("c")
	comp.MustBeNil(err)
	Println(tags)
	recent, err := db.GetRecent()
	comp.MustBeNil(err)
	Println(recent)
	title, err := db.GetTitle(recent[0].ID)
	comp.MustBeNil(err)
	Println(title)
	title = "john"
	titleNotes, err := db.SearchNotesByTitle(title)
	comp.MustBeNil(err)
	Println(titleNotes)
	title = "john questions"
	titleNotes, err = db.SearchNotesByTitle(title)
	comp.MustBeNil(err)
	Println(titleNotes)
}
