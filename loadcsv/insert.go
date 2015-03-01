package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

func insertRecords(table string, cols []string, records [][]string, logger chan<- string) error {
	insertErrors := 0

	db, err := sql.Open("mysql", DB_STRING)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	// create insert statement string
	var queryStringBuf bytes.Buffer
	queryStringBuf.WriteString("INSERT INTO " + table + " ( ")
	for i := 0; i < len(cols); i++ {
		queryStringBuf.WriteString(cols[i])
		if i < len(cols)-1 {
			queryStringBuf.WriteString(", ")
		}
	}
	queryStringBuf.WriteString(" ) VALUES ( ")

	for i := 0; i < len(cols); i++ {
		queryStringBuf.WriteString(DB_PREPARE_RUNE)
		if i < len(cols)-1 {
			queryStringBuf.WriteString(", ")
		}
	}
	queryStringBuf.WriteString(" )")
	query := queryStringBuf.String()

	// create prepared statement
	stmt, err := db.Prepare(query)
	if err != nil {
		logger <- fmt.Sprintf("ERROR: Unable to prepare statement: %s\n", query)
		fmt.Printf("%q\n", err)
		return err
	}

	logger <- fmt.Sprintf("INFO: Using prepared statement \"%s\"", query)

	// watch for dates in parsed data
	dateRegex, _ := regexp.Compile(DATE_REGEX)

	// insert each record
	for recNum, singleRecord := range records {
		// create pointers to parsed data
		record := make([]interface{}, len(singleRecord))
		for i, val := range singleRecord {
			// convert strings to dates if identified
			if dateRegex.MatchString(val) {
				record[i], _ = time.Parse(DATE_FORMAT, val)
			} else {
				record[i] = val
			}
		}
		// execute insert
		res, err := stmt.Exec(record...)
		if err != nil {
			logger <- fmt.Sprintf("ERROR: Unable to insert record #%d %q --- Database returned: \"%s\"", recNum+1, singleRecord, err)
			insertErrors++
		} else {
			id, err := res.LastInsertId()
			if err != nil {
				id = -1 // in case database doesn't support seq ID
			}
			logger <- fmt.Sprintf("INFO: Inserted record ID: %d %q", id, singleRecord)
		}
	}

	if insertErrors > 0 {
		return fmt.Errorf("ERROR: Unable to insert %d records. See log.txt for error details", insertErrors)
	} else {
		return nil
	}
}
