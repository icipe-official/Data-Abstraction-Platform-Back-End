package main

import (
	initdatabase "data_administration_platform/internal/init_database"
	"data_administration_platform/internal/pkg/lib"
	"fmt"
)

func main() {
	const currentSection = "Init Database"
	db, err := lib.OpenDbConnection()
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, err.Error())
	}
	defer db.Close()

	if rowsAffected, err := initdatabase.InsertProjectRoles(db); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, err.Error())
	} else {
		lib.Log(lib.LOG_INFO, currentSection, fmt.Sprintf("%v project roles inserted", rowsAffected))
	}

	if rowsAffected, err := initdatabase.InsertDirectoryIamTicketTypes(db); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, err.Error())
	} else {
		lib.Log(lib.LOG_INFO, currentSection, fmt.Sprintf("%v directory iam ticket types inserted", rowsAffected))
	}

	if rowsAffected, err := initdatabase.InsertStorageTypes(db); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, err.Error())
	} else {
		lib.Log(lib.LOG_INFO, currentSection, fmt.Sprintf("%v storage types inserted", rowsAffected))
	}
}
