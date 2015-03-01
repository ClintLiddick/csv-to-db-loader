package main

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

// const DB_STRING = "golang:golang@tcp(192.168.50.5:3306)/test"
const DB_PREPARE_RUNE = "?"
const DATE_REGEX = "\\d{2}/\\d{2}/\\d{4}"
const DATE_FORMAT = "01/02/2006"

// setup full length flags
var fileFlag = flag.String("file", "upload.csv", "csv file with data to upload")
var tableFlag = flag.String("table", "", "table name to load data into")
var insertFlag = flag.Bool("insert", false, "set true to insert new records")
var updateFlag = flag.Bool("update", false, "set true to updates existing records")

func init() {
	// setup short flag
	flag.StringVar(fileFlag, "f", "upload.csv", "csv file with data to upload")
	flag.StringVar(tableFlag, "t", "", "table name to load data into")
	flag.BoolVar(insertFlag, "i", false, "set true to insert new records")
	flag.BoolVar(updateFlag, "u", false, "set true to updates existing records")
}

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	// validate arguments
	err := validateArgument()

	// read settings
	settings, err := readSettings()
	if err != nil {
		return fmt.Errorf("%s\n", err)
	}

	// open file
	file, err := os.Open(*fileFlag)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to open data file \"%s\"\n", *fileFlag)
	}
	defer file.Close()

	// read csv data
	csvreader := csv.NewReader(file)
	records, err := csvreader.ReadAll()
	if err != nil {
		return fmt.Errorf("ERROR: Unable to read data file \"%s\"\n", *fileFlag)
	}

	// use go-routine to write logfile for performance
	var wg sync.WaitGroup
	defer wg.Wait()
	logchan := logger(&wg)
	defer close(logchan)

	dsn := settings.generateDsnString()
	logchan <- fmt.Sprintf("Connecting to database with DSN: %s", dsn)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("ERROR: Unable to connect to database\n")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("ERROR: Unable to connect to database\n")
	}

	if *insertFlag {
		// insert records
		err = insertRecords(db, *tableFlag, records[0], records[1:], logchan)
		if err != nil {
			return fmt.Errorf("%s\n", err)
		}
		fmt.Println("All records inserted successfully! See log.txt for details")
	} else if *updateFlag {
		err = updateRecords(db, *tableFlag, records[0], records[1:], logchan)
		if err != nil {
			return fmt.Errorf("%s\n", err)
		}
		fmt.Println("All records updated successfully! See log.txt for details")
	}

	return nil

	// done <- true
	// close(done)
}

func validateArgument() error {
	err := false

	if len(*tableFlag) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: Must provide table name with --table flag\n")
		err = true
	}

	if *insertFlag && *updateFlag {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot set --insert and --update. Use separate files.\n")
		err = true
	}

	if !*insertFlag && !*updateFlag {
		fmt.Fprintf(os.Stderr, "ERROR: Must set either --insert or --update\n")
		err = true
	}

	if filepath.Ext(*fileFlag) != ".csv" {
		fmt.Fprintln(os.Stderr, "ERROR: File must be csv\n")
		err = true
	}

	if err {
		return fmt.Errorf("usage")
	} else {
		return nil
	}
}
