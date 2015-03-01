package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

/*
	json format

	{
		"user" : "myusername",
		"pass" : "cleartextpassword",
		"ip" : "192.168.0.1",
		"port" : 3306,
		"database" : "mydbname"
	}
*/

type Settings struct {
	User     string
	Pass     string
	Ip       string
	Port     int
	Database string
}

var settingsFlag = flag.String("settings", "settings.json", "the json file with connection settings")

func init() {
	flag.StringVar(settingsFlag, "s", "settings.json", "the json file with connection settings")
}

func readSettings() (*Settings, error) {
	var settings = new(Settings)
	// open file
	settingsData, err := ioutil.ReadFile(*settingsFlag)
	if err != nil {
		return settings, fmt.Errorf("ERROR: Unable to read file \"%s\"\n", *settingsFlag)
	}

	err = json.Unmarshal(settingsData, settings)
	if err != nil {
		return settings, fmt.Errorf("ERROR: Unable to parse settings data from file \"%s\"\n", *settingsFlag)
	}

	return settings, nil
}

func (s *Settings) generateDsnString() string {
	return s.User + ":" + s.Pass + "@tcp(" + s.Ip + ":" + fmt.Sprintf("%d", s.Port) + ")/" + s.Database
}
