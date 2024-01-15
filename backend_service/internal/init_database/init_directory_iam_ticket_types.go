package initdatabase

import (
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	"data_administration_platform/internal/pkg/lib"
	"database/sql"
	"fmt"
)

var directoryIamTicketTypes = []*model.DirectoryIamTicketTypes{
	{
		ID:          "email_verification",
		Description: "Ticket should be used when a user wants to verify their email",
	},
	{
		ID:          "password_reset",
		Description: "Ticket should be used when a user wants to reset their password",
	},
}

func InsertDirectoryIamTicketTypes(db *sql.DB) (int64, error) {
	insertQuery := table.DirectoryIamTicketTypes.INSERT(table.DirectoryIamTicketTypes.AllColumns).MODELS(directoryIamTicketTypes).ON_CONFLICT(table.DirectoryIamTicketTypes.ID).DO_NOTHING()
	sqlResults, err := insertQuery.Exec(db)
	if err != nil {
		return 0, fmt.Errorf("could not insert directory iam ticket types | reason: %v", err)
	}
	rowsAffected, err := sqlResults.RowsAffected()
	if err != nil {
		lib.Log(lib.LOG_WARNING, "Directory Iam Ticket Types", fmt.Sprintf("Could not determine number of rows inserted for directory iam ticket types | reason: %v", err))
	}
	return rowsAffected, nil
}
