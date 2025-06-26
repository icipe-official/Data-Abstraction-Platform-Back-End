package lib

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/gofrs/uuid/v5"
	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
	intlibmmodel "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib/metadata_model"
)

func ExportToCSV(ctx context.Context, searchResults *intdoment.MetadataModelSearchResults, storageFilesTemporaryRepository intdomint.StorageFilesTemporaryRepository, fileService intdomint.FileService, dataName string) (*intdoment.StorageFilesTemporary, error) {
	extract2DFields, err := intlibmmodel.NewExtract2DFields(searchResults.MetadataModel, true, true, true, nil)
	if err != nil {
		return nil, err
	}
	if err := extract2DFields.Extract(); err != nil {
		return nil, err
	}
	if err := extract2DFields.Reposition(); err != nil {
		return nil, err
	}
	if err := extract2DFields.RemoveSkipped(); err != nil {
		return nil, err
	}

	objectTo2DArray, err := intlibmmodel.NewConvertObjectsTo2DArray(searchResults.MetadataModel, nil, true, true)
	if err != nil {
		return nil, err
	}

	if err := objectTo2DArray.Convert(searchResults.Data); err != nil {
		return nil, err
	}

	rows := make([][]any, 1+len(objectTo2DArray.Array2D()))
	rows[0] = make([]any, len(extract2DFields.Fields()))
	for fIndex, field := range extract2DFields.Fields() {
		rows[0][fIndex] = intlibmmodel.GetFieldGroupName(field, "")
	}

	for rIndex, r := range objectTo2DArray.Array2D() {
		rows[rIndex+1] = r
	}

	tmpFile, err := os.CreateTemp("", "data-*.csv")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	csvWriter := csv.NewWriter(tmpFile)

	for _, r := range rows {
		rString := make([]string, len(r))
		for rdIndex, rd := range r {
			if rdString, ok := rd.(string); ok {
				rString[rdIndex] = rdString
			} else if rdUUID, ok := rd.(uuid.UUID); ok {
				rString[rdIndex] = rdUUID.String()
			} else if rTime, ok := rd.(time.Time); ok {
				rString[rdIndex] = rTime.String()
			} else if rd == nil {
				rString[rdIndex] = ""
			} else {
				rString[rdIndex] = fmt.Sprintf("%v", intlib.JsonStringifyMust(rd))
			}
		}
		if err := csvWriter.Write(rString); err != nil {
			return nil, err
		}
	}

	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		return nil, err
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		return nil, err
	}

	storageFilesTemporary := new(intdoment.StorageFilesTemporary)
	if fileInfo, err := tmpFile.Stat(); err == nil {
		storageFilesTemporary.OriginalName = []string{fmt.Sprintf("%s_export.csv", dataName)}
		storageFilesTemporary.Tags = []string{"tmp"}
		storageFilesTemporary.StorageFileMimeType = []string{"text/csv"}
		storageFilesTemporary.SizeInBytes = []int64{fileInfo.Size()}
	}

	if value, err := storageFilesTemporaryRepository.RepoStorageFilesTemporaryInsertOne(ctx, fileService, storageFilesTemporary, tmpFile, nil); err != nil {
		return nil, err
	} else {
		return value, nil
	}
}
