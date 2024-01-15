package main

import (
	"bufio"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	"data_administration_platform/internal/pkg/lib"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

const currentSection = "Create System User"

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your name:")
	name, err := reader.ReadString('\n')
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not read your name | reason: %v", err))
	}
	name = strings.Trim(name, " \n")
	if len(name) < 3 {
		lib.Log(lib.LOG_FATAL, currentSection, "Invalid name")
	}

	fmt.Print("Enter email:")
	email, err := reader.ReadString('\n')
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not read your email | reason: %v", err))
	}
	email = strings.Trim(email, " \n")
	if len(email) < 3 {
		lib.Log(lib.LOG_FATAL, currentSection, "Invalid email")
	}

	fmt.Print("Enter password (at least 6 characters):")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not read your password | reason: %v", err))
	}
	if len(bytePassword) < 7 {
		lib.Log(lib.LOG_FATAL, currentSection, "Invalid password")
	}

	db, err := lib.OpenDbConnection()
	if err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, err.Error())
	}
	defer db.Close()
	fmt.Println()

	newUser := model.Directory{
		Name:     name,
		Contacts: []string{fmt.Sprintf("email%v%v", lib.OPTS_SPLIT, email)},
	}
	lib.Log(lib.LOG_INFO, currentSection, "Creating new user in directory...")
	insertQuery := table.Directory.INSERT(table.Directory.Name, table.Directory.Contacts).MODEL(newUser).RETURNING(table.Directory.ID)
	if err = insertQuery.Query(db, &newUser); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not create new user in directory | reason: %v", err))
	}

	pass := strings.Trim(string(bytePassword), " \n")
	newIamUser := model.DirectoryIam{
		DirectoryID:     newUser.ID,
		Email:           &email,
		Password:        &pass,
		IsEmailVerified: true,
	}
	lib.Log(lib.LOG_INFO, currentSection, "Creating new user credentials...")
	insertQuery = table.DirectoryIam.
		INSERT(
			table.DirectoryIam.DirectoryID,
			table.DirectoryIam.Email,
			table.DirectoryIam.Password,
			table.DirectoryIam.IsEmailVerified,
		).MODEL(newIamUser).RETURNING(table.DirectoryIam.DirectoryID)
	if err = insertQuery.Query(db, &newIamUser); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not create new user credentials | reason: %v", err))
	}

	newSystemUser := model.DirectorySystemUsers{
		DirectoryID: newIamUser.DirectoryID,
	}
	lib.Log(lib.LOG_INFO, currentSection, "Adding user as a system user...")
	insertQuery = table.DirectorySystemUsers.INSERT(table.DirectorySystemUsers.DirectoryID).MODEL(newSystemUser).RETURNING(table.DirectorySystemUsers.DirectoryID)
	if err = insertQuery.Query(db, &newSystemUser); err != nil {
		lib.Log(lib.LOG_FATAL, currentSection, fmt.Sprintf("Could not add system user | reason: %v", err))
	}
	lib.Log(lib.LOG_INFO, currentSection, fmt.Sprintf("User %v created", newSystemUser.DirectoryID))
}
