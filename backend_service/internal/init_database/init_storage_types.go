package initdatabase

import (
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	"data_administration_platform/internal/pkg/lib"
	"database/sql"
	"fmt"
)

var storageTypes = []*model.StorageTypes{
	{
		ID:         "local",
		Properties: "{\"path\":\"\"}",
	},
	{
		ID:         "azure_blob_mounted",
		Properties: "{\"path\":\"\"}",
	},
}

func InsertStorageTypes(db *sql.DB) (int64, error) {
	insertQuery := table.StorageTypes.INSERT(table.StorageTypes.AllColumns).MODELS(storageTypes).ON_CONFLICT(table.StorageTypes.ID).DO_NOTHING()
	sqlResults, err := insertQuery.Exec(db)
	if err != nil {
		return 0, fmt.Errorf("could not insert storage types | reason: %v", err)
	}
	rowsAffected, err := sqlResults.RowsAffected()
	if err != nil {
		lib.Log(lib.LOG_WARNING, "Storage Types", fmt.Sprintf("Could not determine number of rows inserted for storage types | reason: %v", err))
	}
	return rowsAffected, nil
}
