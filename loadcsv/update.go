package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

// assumes first column is unique ID
func updateRecords(db *sql.DB, table string, cols []string, records [][]string, logger chan<- string) error {
	updateErrors := 0

	// create insert statement string
	var queryStringBuf bytes.Buffer
	queryStringBuf.WriteString("UPDATE " + table + " SET ")
	for i := 1; i < len(cols); i++ {
		queryStringBuf.WriteString(cols[i] + " = " + DB_PREPARE_RUNE)
		if i < len(cols)-1 {
			queryStringBuf.WriteString(", ")
		}
	}
	queryStringBuf.WriteString(" WHERE " + cols[0] + " = ?")
	query := queryStringBuf.String()

	// create prepared statement
	stmt, err := db.Prepare(query)
	if err != nil {
		logger <- fmt.Sprintf("ERROR: Unable to prepare statement: \"%s\" --- Database returned: %s", query, err)
		return fmt.Errorf("ERROR: Unable to prepare statement. See log for details")
	}

	logger <- fmt.Sprintf("INFO: Using prepared statement \"%s\"", query)

	// watch for dates in parsed data
	dateRegex, _ := regexp.Compile(DATE_REGEX)

	// insert each record
	for recNum, singleRecord := range records {
		lastIndex := len(singleRecord) - 1
		// create pointers to parsed data
		preparedValues := make([]interface{}, len(singleRecord))
		// set last field in preparedValues (the WHERE = ? interpolation) = first field in record (ID)
		preparedValues[lastIndex] = singleRecord[0]
		// set data fields in preparedValues (all but last)
		for i := 0; i < len(singleRecord)-1; i++ {
			// offset one since singleRecord[0] is ID used above
			val := singleRecord[i+1]
			// convert strings to dates if identified
			if dateRegex.MatchString(val) {
				preparedValues[i], _ = time.Parse(DATE_FORMAT, val)
			} else {
				preparedValues[i] = val
			}
		}
		// execute insert
		_, err := stmt.Exec(preparedValues...)
		if err != nil {
			logger <- fmt.Sprintf("ERROR: Unable to update record #%d %q --- Database returned: \"%s\"", recNum+1, singleRecord, err)
			updateErrors++
		}
		logger <- fmt.Sprintf("INFO: Updated record %q", singleRecord)
	}

	if updateErrors > 0 {
		return fmt.Errorf("ERROR: Unable to update %d records. See log.txt for error details", updateErrors)
	} else {
		return nil
	}
}
