package db

import (
	"database/sql"
	. "fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/drgrib/alfred-bear/comp"
)

//////////////////////////////////////////////
/// query templates
//////////////////////////////////////////////

const tagQuery = `
	SELECT DISTINCT
		t.ZTITLE 
	FROM 
		ZSFNOTE n 
		INNER JOIN Z_5TAGS nt ON n.Z_PK = nt.Z_5NOTES 
		INNER JOIN ZSFNOTETAG t ON nt.Z_10TAGS = t.Z_PK 
	WHERE 
		n.ZARCHIVED=0 
		AND n.ZTRASHED=0 
		AND lower(t.ZTITLE) LIKE lower('%%%v%%')
	ORDER BY 
		t.ZMODIFICATIONDATE DESC 
	LIMIT %v
`

const recentQuery = `
	SELECT DISTINCT
		ZUNIQUEIDENTIFIER, ZTITLE 
	FROM 
		ZSFNOTE 
	WHERE 
		ZARCHIVED=0 
		AND ZTRASHED=0 
	ORDER BY 
		ZMODIFICATIONDATE DESC 
	LIMIT %v
`

const titleByIDQuery = `
	SELECT DISTINCT
		ZTITLE 
	FROM 
		ZSFNOTE 
	WHERE 
		ZARCHIVED=0 
		AND ZTRASHED=0 
		AND ZUNIQUEIDENTIFIER='%v' 
	ORDER BY 
		ZMODIFICATIONDATE DESC 
	LIMIT %v
`

const notesByTitleQuery = `
	SELECT DISTINCT
		ZUNIQUEIDENTIFIER, ZTITLE 
	FROM 
		ZSFNOTE 
	WHERE 
		ZARCHIVED=0 
		AND ZTRASHED=0 
		AND lower(ZTITLE) LIKE lower('%%%v%%')
	ORDER BY 
		ZMODIFICATIONDATE DESC 
	LIMIT %v
`

//////////////////////////////////////////////
/// Note
//////////////////////////////////////////////

type Note struct {
	ID, Title string
}

//////////////////////////////////////////////
/// BearDB
//////////////////////////////////////////////

type BearDB struct {
	lite  LiteDB
	limit int
}

func NewBearDB() (BearDB, error) {
	path := comp.Expanduser("~/Library/Containers/net.shinyfrog.bear/Data/Documents/Application Data/database.sqlite")
	lite, err := NewLiteDB(path)
	limit := 25
	db := BearDB{lite, limit}
	return db, err
}

func (db BearDB) SearchTags(s string) ([]string, error) {
	q := Sprintf(tagQuery, s, db.limit)
	tags, err := db.lite.QueryStrings(q)
	return tags, err
}

func toNotes(maps []map[string]string) []Note {
	notes := []Note{}
	for _, m := range maps {
		n := Note{
			ID:    m["ZUNIQUEIDENTIFIER"],
			Title: m["ZTITLE"],
		}
		notes = append(notes, n)
	}
	return notes
}

func (db BearDB) GetRecent() ([]Note, error) {
	q := Sprintf(recentQuery, db.limit)
	maps, err := db.lite.QueryStringMaps(q)
	if err != nil {
		return []Note{}, err
	}
	notes := toNotes(maps)
	return notes, err
}

func (db BearDB) GetTitle(id string) (string, error) {
	q := Sprintf(titleByIDQuery, id, db.limit)
	titles, err := db.lite.QueryStrings(q)
	if err != nil {
		return "", err
	}
	if len(titles) == 0 {
		return "", Errorf(
			"No notes for ID '%v'", id)
	}
	return titles[0], err
}

func (db BearDB) simpleSearchByTitle(title string) ([]Note, error) {
	q := Sprintf(notesByTitleQuery, title, db.limit)
	maps, err := db.lite.QueryStringMaps(q)
	if err != nil {
		return []Note{}, err
	}
	notes := toNotes(maps)
	return notes, err
}

func updateNoteMap(m map[Note]bool, items []Note) {
	for _, n := range items {
		m[n] = true
	}
}

func noteMapContains(m map[Note]bool, n Note) bool {
	_, exists := m[n]
	return exists
}

func updateNotes(notes, moreNotes []Note, noteSet map[Note]bool) []Note {
	for _, n := range moreNotes {
		if !noteMapContains(noteSet, n) {
			notes = append(notes, n)
		}
	}
	updateNoteMap(noteSet, moreNotes)
	return notes
}

func (db BearDB) SearchNotesByTitle(title string) ([]Note, error) {
	noteSet := map[Note]bool{}
	notes, err := db.simpleSearchByTitle(title)
	if err != nil {
		return []Note{}, err
	}
	updateNoteMap(noteSet, notes)
	split := strings.Split(title, " ")
	if len(split) > 1 {
		// word gap search
		join := strings.Join(split, "% %")
		moreNotes, err := db.simpleSearchByTitle(join)
		if err != nil {
			return notes, err
		}
		notes = updateNotes(notes, moreNotes, noteSet)
	}
	// letter gap search
	gapList := []string{}
	for _, word := range split {
		gapWord := string(word[0])
		wordLen := len(word)
		if wordLen > 1 {
			last := string(word[wordLen-1])
			gapWord += "%" + last
		}
		gapList = append(gapList, word)
	}
	join := strings.Join(gapList, "% %")
	moreNotes, err := db.simpleSearchByTitle(join)
	if err != nil {
		return notes, err
	}
	notes = updateNotes(notes, moreNotes, noteSet)
	return notes, err
}

//////////////////////////////////////////////
/// LiteDB
//////////////////////////////////////////////

type LiteDB struct {
	db *sql.DB
}

func NewLiteDB(path string) (LiteDB, error) {
	db, err := sql.Open("sqlite3", path)
	lite := LiteDB{db}
	return lite, err
}

func (lite LiteDB) Query(q string) ([]map[string]interface{}, error) {
	results := []map[string]interface{}{}
	rows, err := lite.db.Query(q)
	if err != nil {
		return results, err
	}
	defer rows.Close()
	// credit
	// https://kylewbanks.com/blog/query-result-to-map-in-golang
	cols, err := rows.Columns()
	if err != nil {
		return results, err
	}
	for rows.Next() {
		m := make(map[string]interface{})
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}
		if err := rows.Scan(columnPointers...); err != nil {
			return results, err
		}
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			m[colName] = *val
		}
		results = append(results, m)
	}
	return results, err
}

func B2S(bs []uint8) string {
	// credit
	// https://stackoverflow.com/a/28848879/130427
	b := make([]byte, len(bs))
	for i, v := range bs {
		b[i] = byte(v)
	}
	return string(b)
}

func (lite LiteDB) QueryStringMaps(q string) ([]map[string]string, error) {
	sResults := []map[string]string{}
	iResults, err := lite.Query(q)
	if err != nil {
		return sResults, err
	}
	for _, iMap := range iResults {
		sMap := map[string]string{}
		for k, v := range iMap {
			sMap[k] = B2S(v.([]uint8))
		}
		sResults = append(sResults, sMap)
	}
	return sResults, err
}

func (lite LiteDB) QueryStrings(q string) ([]string, error) {
	sResults := []string{}
	iResults, err := lite.Query(q)
	if err != nil {
		return sResults, err
	}
	for _, iMap := range iResults {
		s := ""
		for _, v := range iMap {
			s = B2S(v.([]uint8))
		}
		sResults = append(sResults, s)
	}
	return sResults, err
}
