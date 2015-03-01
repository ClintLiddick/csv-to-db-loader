package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

const DB_STRING = "golang:golang@tcp(192.168.50.5:3306)/test"
const DB_PREPARE_RUNE = "?"
const TABLE = "poles"
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
	flag.StringVar(tableFlag, "table", "", "table name to load data into")
	flag.BoolVar(insertFlag, "i", false, "set true to insert new records")
	flag.BoolVar(updateFlag, "u", false, "set true to updates existing records")
}

func main() {
	flag.Parse()

	// validate arguments
	err := validateArgument()

	// open file
	file, err := os.Open(*fileFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to open file \"%s\"\n", *fileFlag)
		os.Exit(1)
	}
	defer file.Close()

	// read csv data
	csvreader := csv.NewReader(file)
	records, err := csvreader.ReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Unable to read file \"%s\"\n", *fileFlag)
		os.Exit(1)
	}

	// use go-routine to write logfile for performance
	done := make(chan bool)
	logchan := logger(done)

	// insert records
	err = insertRecords(TABLE, records[0], records[1:], logchan)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}

	done <- true
	close(done)

}

func validateArgument() error {
	err := false

	if len(*tableFlag) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: Must provide table name with --table flag\n")
		err = true
	}

	if *insertFlag && *updateFlag {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot set --insert=true and --update=true. Use separate files.\n")
		err = true
	}

	if !*insertFlag || *updateFlag {
		fmt.Fprintf(os.Stderr, "ERROR: Must set either --insert=true or --update=true\n")
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
