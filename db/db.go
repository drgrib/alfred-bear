package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

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

func (lite LiteDB) StringQuery(q string) ([]map[string]string, error) {
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
