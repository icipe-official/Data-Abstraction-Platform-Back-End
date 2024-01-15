package abstractions

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"encoding/csv"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

const currentSection = "Abstractions"

type fileFromAbstractionData struct {
	SingleSheet       string                 `json:"-"`
	FileType          string                 `json:"-"`
	ModelTemplate     map[string]interface{} `json:"-"`
	FilePath          string
	FileContentType   string
	File              *os.File
	CsvWriter         *csv.Writer
	ExcelWorkbook     *excelize.File
	ExcelStreamWriter *excelize.StreamWriter
	HeadersWritten    bool
	CurrentRow        int64
}

type abstractions struct {
	FileFromData            fileFromAbstractionData
	FileFromAbstractionData struct {
		ModelTemplate string
	}
	ModelTemplate            model.ModelTemplates
	AbstractionReview        abstractionReview
	CreatedOnGreaterThan     string `json:"-"`
	CreatedOnLessThan        string `json:"-"`
	LastUpdatedOnGreaterThan string `json:"-"`
	LastUpdatedOnLessThan    string `json:"-"`
	IsVerified               string `json:"-"`
	Limit                    int    `json:"-"`
	Offset                   int    `json:"-"`
	SortyBy                  string `json:"-"`
	SortByOrder              string `json:"-"`
	ModelTemplateID          uuid.UUID
	DirectoryID              uuid.UUID
	ProjectID                uuid.UUID
	AbstractionID            uuid.UUID
	AbstractionUpdate        struct {
		Abstraction model.Abstractions
		Columns     []string
	}
	CurrentUser          lib.User
	AbstractionsCreation abstractionCreation
	RetrieveAbstraction  RetrieveOneAbstraction
	RetrieveAbstractions []RetrieveAbstraction
}

type abstractionReview struct {
	model.AbstractionReviews
	Comment model.AbstractionReviewsComments
}

type RetrieveOneAbstraction struct {
	ID                          uuid.UUID `sql:"primary_key"`
	ModelTemplateID             uuid.UUID
	ModelTemplate               model.ModelTemplates
	FileID                      uuid.UUID
	FileContentType             string
	FileTags                    string
	AbstractorDirectoryID       uuid.UUID
	AbstractorDirectoryName     string
	AbstractorDirectoryContacts pq.StringArray
	ProjectID                   uuid.UUID
	Tags                        *string
	Abstraction                 string
	IsVerified                  bool
	CreatedOn                   time.Time
	LastUpdatedOn               time.Time
	AbstractionReviews          []AbstractionReview
	AbstractionReviewsComments  []AbstractionReviewComment
}

type RetrieveAbstraction struct {
	ID                          uuid.UUID `sql:"primary_key"`
	ModelTemplateID             uuid.UUID
	FileID                      uuid.UUID
	FileContentType             string
	FileTags                    string
	AbstractorDirectoryID       uuid.UUID
	AbstractorDirectoryName     string
	AbstractorDirectoryContacts pq.StringArray
	ProjectID                   uuid.UUID
	Tags                        *string
	Abstraction                 string
	IsVerified                  bool
	CreatedOn                   time.Time
	LastUpdatedOn               time.Time
	AbstractionReviews          []AbstractionReview
	AbstractionReviewsComments  []AbstractionReviewComment
}

type AbstractionReview struct {
	ReviewerDirectoryID       uuid.UUID `sql:"primary_key"`
	ReviewerDirectoryName     string
	ReviewerDirectoryContacts pq.StringArray
	Review                    bool
	ReviewCreatedOn           time.Time
	ReviewLastUpdatedOn       time.Time
}

type AbstractionReviewComment struct {
	ID                         uuid.UUID `sql:"primary_key"`
	CommenterDirectoryID       uuid.UUID
	CommenterDirectoryName     string
	CommenterDirectoryContacts pq.StringArray
	Comment                    string
	CommentCreatedOn           time.Time
}

type abstractionCreation struct {
	DirectoryID               uuid.UUID
	ProjectID                 uuid.UUID
	ModelTemplateID           uuid.UUID
	Files                     []model.Files
	FilesSearchQuery          string
	Tags                      string
	SkipFilesWithAbstractions bool
}
