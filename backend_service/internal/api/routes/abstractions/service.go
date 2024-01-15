package abstractions

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"os"
	"strings"
	"time"

	jet "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

func (n *abstractions) genFileFromAbstractions() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	n.Limit = 1000
	n.Offset = 0

	whereCondition := table.Abstractions.ModelTemplateID.EQ(jet.UUID(n.ModelTemplateID)).AND(table.Abstractions.ProjectID.EQ(jet.UUID(n.ProjectID)))
	if n.CreatedOnGreaterThan != "" {
		whereCondition = whereCondition.AND(table.Abstractions.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD"))))
	}
	if n.CreatedOnLessThan != "" {
		whereCondition = whereCondition.AND(table.Abstractions.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD"))))
	}
	if n.LastUpdatedOnGreaterThan != "" {
		whereCondition = whereCondition.AND(table.Abstractions.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD"))))
	}
	if n.LastUpdatedOnLessThan != "" {
		whereCondition = whereCondition.AND(table.Abstractions.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD"))))
	}
	if n.IsVerified != "" {
		var iaCondition jet.BoolExpression
		if n.IsVerified == "true" {
			iaCondition = table.Abstractions.IsVerified.IS_TRUE()
		} else {
			iaCondition = table.Abstractions.IsVerified.IS_FALSE()
		}
		whereCondition = whereCondition.AND(iaCondition)
	}

	if n.DirectoryID != uuid.Nil {
		whereCondition = whereCondition.AND(table.Abstractions.DirectoryID.EQ(jet.UUID(n.DirectoryID)))
	}
	abstractionsSelectQuery := jet.SELECT(
		table.Abstractions.ID.AS("retrieve_abstraction.id"),
		table.Abstractions.ModelTemplateID.AS("retrieve_abstraction.model_template_id"),
		table.Abstractions.FileID.AS("retrieve_abstraction.file_id"),
		table.Abstractions.DirectoryID.AS("retrieve_abstraction.abstractor_directory_id"),
		table.Abstractions.ProjectID.AS("retrieve_abstraction.project_id"),
		table.Abstractions.Tags.AS("retrieve_abstraction.tags"),
		table.Abstractions.Abstraction.AS("retrieve_abstraction.abstraction"),
		table.Abstractions.IsVerified.AS("retrieve_abstraction.is_verified"),
		table.Abstractions.CreatedOn.AS("retrieve_abstraction.created_on"),
		table.Abstractions.LastUpdatedOn.AS("retrieve_abstraction.last_updated_on"),
	).FROM(table.Abstractions).WHERE(whereCondition)
	if n.SortyBy != "" {
		switch n.SortyBy {
		case table.Abstractions.CreatedOn.Name():
			if n.SortByOrder == "asc" {
				abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.CreatedOn.ASC())
			} else {
				abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.CreatedOn.DESC())
			}
		case table.Abstractions.LastUpdatedOn.Name():
			if n.SortByOrder == "asc" {
				abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.LastUpdatedOn.ASC())
			} else {
				abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.LastUpdatedOn.DESC())
			}
		}
	}
	selectedAbstractions := abstractionsSelectQuery.AsTable(table.Abstractions.TableName())
	selectQuery := jet.SELECT(
		selectedAbstractions.AllColumns(),
		table.Files.AS("abstractions_files").ContentType.AS("retrieve_abstraction.file_content_type"),
		table.Files.AS("abstractions_files").Tags.AS("retrieve_abstraction.file_tags"),
		table.Directory.AS("abstractions_directory").Name.AS("retrieve_abstraction.abstractor_directory_name"),
		table.Directory.AS("abstractions_directory").Contacts.AS("retrieve_abstraction.abstractor_directory_contacts"),
		table.AbstractionReviews.DirectoryID.AS("abstraction_review.reviewer_directory_id"),
		table.Directory.AS("abstraction_reviews_directory").Name.AS("abstraction_review.reviewer_directory_name"),
		table.Directory.AS("abstraction_reviews_directory").Contacts.AS("abstraction_review.reviewer_directory_contacts"),
		table.AbstractionReviews.Review.AS("abstraction_review.review"),
		table.AbstractionReviews.CreatedOn.AS("abstraction_review.review_created_on"),
		table.AbstractionReviews.LastUpdatedOn.AS("abstraction_review.review_last_updated_on"),
	).FROM(
		selectedAbstractions.
			INNER_JOIN(table.Directory.AS("abstractions_directory"), table.Directory.AS("abstractions_directory").ID.EQ(jet.StringColumn("retrieve_abstraction.abstractor_directory_id").From(selectedAbstractions))).
			INNER_JOIN(table.Files.AS("abstractions_files"), table.Files.AS("abstractions_files").ID.EQ(jet.StringColumn("retrieve_abstraction.file_id").From(selectedAbstractions))).
			LEFT_JOIN(
				table.AbstractionReviews.INNER_JOIN(
					table.Directory.AS("abstraction_reviews_directory"),
					table.AbstractionReviews.DirectoryID.EQ(table.Directory.AS("abstraction_reviews_directory").ID),
				),
				table.AbstractionReviews.AbstractionID.EQ(jet.StringColumn("retrieve_abstraction.id").From(selectedAbstractions)),
			),
	)
	defer func() {
		switch n.FileFromData.FileType {
		case lib.GEN_FILE_CSV:
			if n.FileFromData.CsvWriter != nil {
				n.FileFromData.CsvWriter.Flush()
			}
		case lib.GEN_FILE_EXCEL:
			if n.FileFromData.ExcelStreamWriter != nil {
				if err := n.FileFromData.ExcelStreamWriter.Flush(); err != nil {
					intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Flush excel temp file failed | reason: %v", err))
				}
			}
		}
		if n.FileFromData.FileType == lib.GEN_FILE_EXCEL && n.FileFromData.ExcelWorkbook != nil {
			if err := n.FileFromData.ExcelWorkbook.Close(); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Close excel temp file failed | reason: %v", err))
			}
		}
		if n.FileFromData.File != nil {
			if err := n.FileFromData.File.Close(); err != nil {
				intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Close temp file failed | reason: %v", err))
			}
		}
	}()
	for {
		n.RetrieveAbstractions = []RetrieveAbstraction{}
		if n.Limit > 0 {
			abstractionsSelectQuery = abstractionsSelectQuery.LIMIT(int64(n.Limit))
		}
		if n.Offset > 0 {
			abstractionsSelectQuery = abstractionsSelectQuery.OFFSET(int64(n.Offset))
		}
		if err := selectQuery.Query(db, &n.RetrieveAbstractions); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get abstractions by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get abstractions")
		}
		if len(n.RetrieveAbstractions) == 0 {
			break
		}
		switch n.FileFromData.FileType {
		case lib.GEN_FILE_CSV:
			if n.FileFromData.FilePath == "" {
				n.FileFromData.FileContentType = "text/csv"
				n.FileFromData.FilePath = fmt.Sprintf("%v/%v_%v.csv", lib.TMP_DIR, time.Now().Unix(), uuid.New())
				if n.FileFromData.File, err = os.Create(n.FileFromData.FilePath); err != nil {
					return fmt.Errorf("create new csv file failed | reason: %v", err)
				}
				n.FileFromData.CsvWriter = csv.NewWriter(n.FileFromData.File)
			}
		case lib.GEN_FILE_EXCEL:
			if n.FileFromData.ExcelWorkbook == nil {
				n.FileFromData.ExcelWorkbook = excelize.NewFile()
			}
			if n.FileFromData.ExcelStreamWriter == nil {
				if n.FileFromData.ExcelStreamWriter, err = n.FileFromData.ExcelWorkbook.NewStreamWriter("Sheet1"); err != nil {
					return fmt.Errorf("create stream writer for workbook failed | reason: %v", err)
				}
			}
		}
		if !n.FileFromData.HeadersWritten {
			switch n.FileFromData.FileType {
			case lib.GEN_FILE_CSV:
				n.FileFromData.CsvWriter.Write(lib.GetColumnHeadersForTwoDimensionArray(n.FileFromData.ModelTemplate))
			case lib.GEN_FILE_EXCEL:
				n.FileFromData.CurrentRow = 1
				excelColumnHeaders := []interface{}{}
				for _, ch := range lib.GetColumnHeadersForTwoDimensionArray(n.FileFromData.ModelTemplate) {
					excelColumnHeaders = append(excelColumnHeaders, excelize.Cell{Value: ch})
				}
				if err := n.FileFromData.ExcelStreamWriter.SetRow(fmt.Sprintf("A%v", n.FileFromData.CurrentRow), excelColumnHeaders); err != nil {
					return fmt.Errorf("set column headers in excel sheet of new workbook failed | reason: %v", err)
				}
			}
			n.FileFromData.HeadersWritten = true
		}
		for index := range n.RetrieveAbstractions {
			abstraction := map[string]interface{}{}
			if err := json.Unmarshal([]byte(n.RetrieveAbstractions[index].Abstraction), &abstraction); err != nil {
				continue
			}
			isCompleted := false
			abstractionReviews := []map[string]interface{}{}
			for _, ar := range n.RetrieveAbstractions[index].AbstractionReviews {
				if ar.ReviewerDirectoryID == n.RetrieveAbstractions[index].AbstractorDirectoryID && ar.Review {
					isCompleted = true
				}
				newAbstractionReview := map[string]interface{}{}
				newAbstractionReview["name"] = ar.ReviewerDirectoryName
				newAbstractionReviewContacts := []string{}
				for _, c := range ar.ReviewerDirectoryContacts {
					newAbstractionReviewContacts = append(newAbstractionReviewContacts, fmt.Sprintf("%v-%v", strings.Split(c, intpkglib.OPTS_SPLIT)[0], strings.Split(c, intpkglib.OPTS_SPLIT)[1]))
				}
				newAbstractionReview["contacts"] = strings.Join(newAbstractionReviewContacts, ",")
				if ar.Review {
					newAbstractionReview["review"] = "positive"
				} else {
					newAbstractionReview["review"] = "negative"
				}
				newAbstractionReview["last_updated_on"] = ar.ReviewLastUpdatedOn.Local().Format(time.RFC822)
				abstractionReviews = append(abstractionReviews, newAbstractionReview)
			}
			abstractionInfo := map[string]interface{}{
				"abstraction_id":  n.RetrieveAbstractions[index].ID,
				"abstractor_name": n.RetrieveAbstractions[index].AbstractorDirectoryName,
				"file_id":         n.RetrieveAbstractions[index].FileID,
				"file_tags":       n.RetrieveAbstractions[index].FileTags,
			}
			if isCompleted {
				abstractionInfo["completed"] = "yes"
			} else {
				abstractionInfo["completed"] = "no"
			}
			if n.RetrieveAbstractions[index].IsVerified {
				abstractionInfo["verified"] = "yes"
			} else {
				abstractionInfo["verified"] = "no"
			}
			newAbstraction := map[string]interface{}{
				"abstraction_info": abstractionInfo,
			}
			if len(abstractionReviews) > 0 {
				newAbstraction["abstraction_reviews"] = abstractionReviews
			}
			maps.Copy(newAbstraction, abstraction)
			for _, row := range lib.ConvertMapIntoTwoDimensionArray([][]any{{}}, n.FileFromData.ModelTemplate, newAbstraction, []int{}) {
				switch n.FileFromData.FileType {
				case lib.GEN_FILE_CSV:
					newRow := []string{}
					for _, rd := range row {
						if rd != nil {
							newRow = append(newRow, fmt.Sprintf("%v", rd))
						} else {
							newRow = append(newRow, "")
						}
					}
					n.FileFromData.CsvWriter.Write(newRow)
				case lib.GEN_FILE_EXCEL:
					n.FileFromData.CurrentRow += 1
					if err := n.FileFromData.ExcelStreamWriter.SetRow(fmt.Sprintf("A%v", n.FileFromData.CurrentRow), row); err != nil {
						return fmt.Errorf("set column headers in excel sheet of new workbook failed | reason: %v", err)
					}
				}
			}
		}
		switch n.FileFromData.FileType {
		case lib.GEN_FILE_CSV:
			n.FileFromData.CsvWriter.Flush()
		case lib.GEN_FILE_EXCEL:
			if err := n.FileFromData.ExcelStreamWriter.Flush(); err != nil {
				return fmt.Errorf("save excel buffer content failed | reason: %v", err)
			}
		}
		if len(n.RetrieveAbstractions) < 1000 {
			if n.FileFromData.FileType == lib.GEN_FILE_EXCEL {
				n.FileFromData.FileContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
				n.FileFromData.FilePath = fmt.Sprintf("%v/%v_%v.xlsx", lib.TMP_DIR, time.Now().UTC(), uuid.New())
				if err := n.FileFromData.ExcelWorkbook.SaveAs(n.FileFromData.FilePath); err != nil {
					return fmt.Errorf("save new workbook failed | reason: %v", err)
				}
			}
			break
		} else {
			n.Offset = n.Limit
			n.Limit = n.Limit + 1000
		}
	}

	return nil
}

func (n *abstractions) reviewAbstraction() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	n.AbstractionReview.DirectoryID = n.CurrentUser.DirectoryID

	upsertQuery := table.AbstractionReviews.
		INSERT(
			table.AbstractionReviews.AbstractionID,
			table.AbstractionReviews.DirectoryID,
			table.AbstractionReviews.Review,
		).MODEL(n.AbstractionReview).
		ON_CONFLICT(table.AbstractionReviews.AbstractionID, table.AbstractionReviews.DirectoryID).
		DO_UPDATE(
			jet.SET(table.AbstractionReviews.Review.SET(jet.Bool(n.AbstractionReview.Review))).
				WHERE(
					table.AbstractionReviews.AbstractionID.EQ(jet.UUID(n.AbstractionReview.AbstractionID)).
						AND(table.AbstractionReviews.DirectoryID.EQ(jet.UUID(n.AbstractionReview.DirectoryID))),
				),
		).RETURNING(table.AbstractionReviews.AbstractionID, table.AbstractionReviews.DirectoryID, table.AbstractionReviews.LastUpdatedOn)
	comment := n.AbstractionReview.Comment.Comment
	if err = upsertQuery.Query(db, &n.AbstractionReview); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Review abstraction %v by %v failed | reason: %v", n.AbstractionReview.AbstractionID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Post abstraction review failed")
	}

	positiveReviewsCount := 0
	selectPositiveReviewsQuery := fmt.Sprintf(
		"SELECT COUNT('*') AS review_verified_count FROM public.abstraction_reviews WHERE abstraction_id = '%v' AND review = %v;",
		n.AbstractionReview.AbstractionID,
		true,
	)
	if err = db.QueryRow(selectPositiveReviewsQuery).Scan(&positiveReviewsCount); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get positive reviews for abstraction %v by %v failed | reason: %v", n.AbstractionReview.AbstractionID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Get positive reviewes for abstraction failed")
	}

	n.ModelTemplate = model.ModelTemplates{}
	selectTemplateQuorumQuery := jet.SELECT(
		table.ModelTemplates.VerificationQuorum,
	).FROM(
		table.Abstractions.RIGHT_JOIN(table.ModelTemplates, table.Abstractions.ModelTemplateID.EQ(table.ModelTemplates.ID)),
	).WHERE(table.Abstractions.ID.EQ(jet.UUID(n.AbstractionReview.AbstractionID)))
	if err := selectTemplateQuorumQuery.Query(db, &n.ModelTemplate); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get template quorum for abstraction %v by %v failed | reason: %v", n.AbstractionReview.AbstractionID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Get template quorum failed")
	}
	isAbstractionVerified := false
	if positiveReviewsCount > int(n.ModelTemplate.VerificationQuorum) {
		isAbstractionVerified = true
	}

	abstractionUpdate := model.Abstractions{
		IsVerified: isAbstractionVerified,
	}
	updateAbstractionVerificationQuery := table.Abstractions.
		UPDATE(table.Abstractions.IsVerified).
		MODEL(abstractionUpdate).
		WHERE(table.Abstractions.ID.EQ(jet.UUID(n.AbstractionReview.AbstractionID)))
	if _, err := updateAbstractionVerificationQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update verification for abstraction %v by %v failed | reason: %v", n.AbstractionReview.AbstractionID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Update verification for abstraction failed")
	}

	if len(comment) > 0 {
		n.AbstractionReview.Comment.Comment = comment
		n.AbstractionReview.Comment.DirectoryID = n.CurrentUser.DirectoryID
		n.AbstractionReview.Comment.AbstractionID = n.AbstractionReview.AbstractionID
		insertQuery := table.AbstractionReviewsComments.
			INSERT(
				table.AbstractionReviewsComments.AbstractionID,
				table.AbstractionReviews.DirectoryID,
				table.AbstractionReviewsComments.Comment,
			).MODEL(n.AbstractionReview.Comment)
		if _, err := insertQuery.Exec(db); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Insert review comment for abstraction %v by %v failed | reason: %v", n.AbstractionReview.AbstractionID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Create review comment failed")
		}
	}

	return nil
}

func (n *abstractions) getAbstractions() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	if n.AbstractionID != uuid.Nil {
		n.RetrieveAbstraction = RetrieveOneAbstraction{}

		selectQuery := jet.SELECT(
			table.Abstractions.ID.AS("retrieve_one_abstraction.id"),
			table.Abstractions.ModelTemplateID.AS("retrieve_one_abstraction.model_template_id"),
			table.ModelTemplates.AllColumns,
			table.Abstractions.FileID.AS("retrieve_one_abstraction.file_id"),
			table.Abstractions.DirectoryID.AS("retrieve_one_abstraction.abstractor_directory_id"),
			table.Abstractions.ProjectID.AS("retrieve_one_abstraction.project_id"),
			table.Abstractions.Tags.AS("retrieve_one_abstraction.tags"),
			table.Abstractions.Abstraction.AS("retrieve_one_abstraction.abstraction"),
			table.Abstractions.IsVerified.AS("retrieve_one_abstraction.is_verified"),
			table.Abstractions.CreatedOn.AS("retrieve_one_abstraction.created_on"),
			table.Abstractions.LastUpdatedOn.AS("retrieve_one_abstraction.last_updated_on"),
			table.Files.AS("abstractions_files").ContentType.AS("retrieve_one_abstraction.file_content_type"),
			table.Files.AS("abstractions_files").Tags.AS("retrieve_one_abstraction.file_tags"),
			table.Directory.AS("abstractions_directory").Name.AS("retrieve_one_abstraction.abstractor_directory_name"),
			table.Directory.AS("abstractions_directory").Contacts.AS("retrieve_one_abstraction.abstractor_directory_contacts"),
			table.AbstractionReviews.DirectoryID.AS("abstraction_review.reviewer_directory_id"),
			table.Directory.AS("abstraction_reviews_directory").Name.AS("abstraction_review.reviewer_directory_name"),
			table.Directory.AS("abstraction_reviews_directory").Contacts.AS("abstraction_review.reviewer_directory_contacts"),
			table.AbstractionReviews.Review.AS("abstraction_review.review"),
			table.AbstractionReviews.CreatedOn.AS("abstraction_review.review_created_on"),
			table.AbstractionReviews.LastUpdatedOn.AS("abstraction_review.review_last_updated_on"),
			table.AbstractionReviewsComments.ID.AS("abstraction_review_comment.id"),
			table.AbstractionReviewsComments.DirectoryID.AS("abstraction_review_comment.commenter_directory_id"),
			table.Directory.AS("abstraction_reviews_comments_directory").Name.AS("abstraction_review_comment.commenter_directory_name"),
			table.Directory.AS("abstraction_reviews_comments_directory").Contacts.AS("abstraction_review_comment.commenter_directory_contacts"),
			table.AbstractionReviewsComments.Comment.AS("abstraction_review_comment.comment"),
			table.AbstractionReviewsComments.CreatedOn.AS("abstraction_review_comment.comment_created_on"),
		).FROM(
			table.Abstractions.
				INNER_JOIN(table.ModelTemplates, table.Abstractions.ModelTemplateID.EQ(table.ModelTemplates.ID)).
				INNER_JOIN(table.Directory.AS("abstractions_directory"), table.Directory.AS("abstractions_directory").ID.EQ(table.Abstractions.DirectoryID)).
				INNER_JOIN(table.Files.AS("abstractions_files"), table.Files.AS("abstractions_files").ID.EQ(table.Abstractions.FileID)).
				LEFT_JOIN(
					table.AbstractionReviews.INNER_JOIN(
						table.Directory.AS("abstraction_reviews_directory"),
						table.AbstractionReviews.DirectoryID.EQ(table.Directory.AS("abstraction_reviews_directory").ID),
					),
					table.AbstractionReviews.AbstractionID.EQ(table.Abstractions.ID),
				).
				LEFT_JOIN(
					table.AbstractionReviewsComments.INNER_JOIN(
						table.Directory.AS("abstraction_reviews_comments_directory"),
						table.AbstractionReviewsComments.DirectoryID.EQ(table.Directory.AS("abstraction_reviews_comments_directory").ID),
					),
					table.AbstractionReviewsComments.AbstractionID.EQ(table.Abstractions.ID),
				),
		).WHERE(table.Abstractions.ID.EQ(jet.UUID(n.AbstractionID)))
		if err := selectQuery.Query(db, &n.RetrieveAbstraction); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get abstraction %v by %v failed | reason: %v", n.AbstractionID, n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get abstraction")
		}
	} else {
		whereCondition := table.Abstractions.ModelTemplateID.EQ(jet.UUID(n.ModelTemplateID)).AND(table.Abstractions.ProjectID.EQ(jet.UUID(n.ProjectID)))
		if n.CreatedOnGreaterThan != "" {
			whereCondition = whereCondition.AND(table.Abstractions.CreatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnGreaterThan), jet.String("YYYY-MM-DD"))))
		}
		if n.CreatedOnLessThan != "" {
			whereCondition = whereCondition.AND(table.Abstractions.CreatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.CreatedOnLessThan), jet.String("YYYY-MM-DD"))))
		}
		if n.LastUpdatedOnGreaterThan != "" {
			whereCondition = whereCondition.AND(table.Abstractions.LastUpdatedOn.GT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnGreaterThan), jet.String("YYYY-MM-DD"))))
		}
		if n.LastUpdatedOnLessThan != "" {
			whereCondition = whereCondition.AND(table.Abstractions.LastUpdatedOn.LT_EQ(jet.TO_TIMESTAMP(jet.String(n.LastUpdatedOnLessThan), jet.String("YYYY-MM-DD"))))
		}
		if n.IsVerified != "" {
			var iaCondition jet.BoolExpression
			if n.IsVerified == "true" {
				iaCondition = table.Abstractions.IsVerified.IS_TRUE()
			} else {
				iaCondition = table.Abstractions.IsVerified.IS_FALSE()
			}
			whereCondition = whereCondition.AND(iaCondition)
		}
		if n.DirectoryID != uuid.Nil {
			whereCondition = whereCondition.AND(table.Abstractions.DirectoryID.EQ(jet.UUID(n.DirectoryID)))
		}

		n.RetrieveAbstractions = []RetrieveAbstraction{}
		abstractionsSelectQuery := jet.SELECT(
			table.Abstractions.ID.AS("retrieve_abstraction.id"),
			table.Abstractions.ModelTemplateID.AS("retrieve_abstraction.model_template_id"),
			table.Abstractions.FileID.AS("retrieve_abstraction.file_id"),
			table.Abstractions.DirectoryID.AS("retrieve_abstraction.abstractor_directory_id"),
			table.Abstractions.ProjectID.AS("retrieve_abstraction.project_id"),
			table.Abstractions.Tags.AS("retrieve_abstraction.tags"),
			table.Abstractions.Abstraction.AS("retrieve_abstraction.abstraction"),
			table.Abstractions.IsVerified.AS("retrieve_abstraction.is_verified"),
			table.Abstractions.CreatedOn.AS("retrieve_abstraction.created_on"),
			table.Abstractions.LastUpdatedOn.AS("retrieve_abstraction.last_updated_on"),
		).FROM(table.Abstractions).WHERE(whereCondition)
		if n.Limit > 0 {
			abstractionsSelectQuery = abstractionsSelectQuery.LIMIT(int64(n.Limit))
		}
		if n.Offset > 0 {
			abstractionsSelectQuery = abstractionsSelectQuery.OFFSET(int64(n.Offset))
		}
		if n.SortyBy != "" {
			switch n.SortyBy {
			case table.Abstractions.CreatedOn.Name():
				if n.SortByOrder == "asc" {
					abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.CreatedOn.ASC())
				} else {
					abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.CreatedOn.DESC())
				}
			case table.Abstractions.LastUpdatedOn.Name():
				if n.SortByOrder == "asc" {
					abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.LastUpdatedOn.ASC())
				} else {
					abstractionsSelectQuery = abstractionsSelectQuery.ORDER_BY(table.Abstractions.LastUpdatedOn.DESC())
				}
			}
		}
		selectedAbstractions := abstractionsSelectQuery.AsTable(table.Abstractions.TableName())
		selectQuery := jet.SELECT(
			selectedAbstractions.AllColumns(),
			table.Files.AS("abstractions_files").ContentType.AS("retrieve_abstraction.file_content_type"),
			table.Files.AS("abstractions_files").Tags.AS("retrieve_abstraction.file_tags"),
			table.Directory.AS("abstractions_directory").Name.AS("retrieve_abstraction.abstractor_directory_name"),
			table.Directory.AS("abstractions_directory").Contacts.AS("retrieve_abstraction.abstractor_directory_contacts"),
			table.AbstractionReviews.DirectoryID.AS("abstraction_review.reviewer_directory_id"),
			table.Directory.AS("abstraction_reviews_directory").Name.AS("abstraction_review.reviewer_directory_name"),
			table.Directory.AS("abstraction_reviews_directory").Contacts.AS("abstraction_review.reviewer_directory_contacts"),
			table.AbstractionReviews.Review.AS("abstraction_review.review"),
			table.AbstractionReviews.CreatedOn.AS("abstraction_review.review_created_on"),
			table.AbstractionReviews.LastUpdatedOn.AS("abstraction_review.review_last_updated_on"),
			table.AbstractionReviewsComments.ID.AS("abstraction_review_comment.id"),
			table.AbstractionReviewsComments.DirectoryID.AS("abstraction_review_comment.commenter_directory_id"),
			table.Directory.AS("abstraction_reviews_comments_directory").Name.AS("abstraction_review_comment.commenter_directory_name"),
			table.Directory.AS("abstraction_reviews_comments_directory").Contacts.AS("abstraction_review_comment.commenter_directory_contacts"),
			table.AbstractionReviewsComments.Comment.AS("abstraction_review_comment.comment"),
			table.AbstractionReviewsComments.CreatedOn.AS("abstraction_review_comment.comment_created_on"),
		).FROM(
			selectedAbstractions.
				INNER_JOIN(table.Directory.AS("abstractions_directory"), table.Directory.AS("abstractions_directory").ID.EQ(jet.StringColumn("retrieve_abstraction.abstractor_directory_id").From(selectedAbstractions))).
				INNER_JOIN(table.Files.AS("abstractions_files"), table.Files.AS("abstractions_files").ID.EQ(jet.StringColumn("retrieve_abstraction.file_id").From(selectedAbstractions))).
				LEFT_JOIN(
					table.AbstractionReviews.INNER_JOIN(
						table.Directory.AS("abstraction_reviews_directory"),
						table.AbstractionReviews.DirectoryID.EQ(table.Directory.AS("abstraction_reviews_directory").ID),
					),
					table.AbstractionReviews.AbstractionID.EQ(jet.StringColumn("retrieve_abstraction.id").From(selectedAbstractions)),
				).
				LEFT_JOIN(
					table.AbstractionReviewsComments.INNER_JOIN(
						table.Directory.AS("abstraction_reviews_comments_directory"),
						table.AbstractionReviewsComments.DirectoryID.EQ(table.Directory.AS("abstraction_reviews_comments_directory").ID),
					),
					table.AbstractionReviewsComments.AbstractionID.EQ(jet.StringColumn("retrieve_abstraction.id").From(selectedAbstractions)),
				),
		)
		if err := selectQuery.Query(db, &n.RetrieveAbstractions); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get abstractions by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return lib.NewError(http.StatusInternalServerError, "Could not get abstractions")
		}
	}

	return nil
}

func (n *abstractions) deleteAbstraction() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	deleteQuery := table.Abstractions.DELETE().WHERE(table.Abstractions.ID.EQ(jet.UUID(n.AbstractionID)).AND(table.Abstractions.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))))

	if sqlResults, err := deleteQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Delete abstraction %v by %v failed | reason: %v", n.AbstractionID, n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not delete abstraction")
	} else {
		if deletedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of abstractions deleted failed while deleting %v | reason: %v", n.AbstractionID, err))
			return -1, nil
		} else {
			return deletedRows, err
		}
	}
}

func (n *abstractions) updateAbstraction() error {
	columnsToUpdate := make(jet.ColumnList, 0)
	for _, column := range n.AbstractionUpdate.Columns {
		switch column {
		case table.Abstractions.Tags.Name():
			if len(*n.AbstractionUpdate.Abstraction.Tags) >= 3 {
				columnsToUpdate = append(columnsToUpdate, table.Abstractions.Tags)
			}
		case table.Abstractions.Abstraction.Name():
			columnsToUpdate = append(columnsToUpdate, table.Abstractions.Abstraction)
		}
	}

	if len(columnsToUpdate) < 1 {
		return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	updateQuery := table.Abstractions.
		UPDATE(columnsToUpdate).
		MODEL(n.AbstractionUpdate.Abstraction).
		WHERE(
			table.Abstractions.ID.EQ(jet.UUID(n.AbstractionID)).
				AND(table.Abstractions.ProjectID.EQ(jet.UUID(n.ProjectID))).
				AND(table.Abstractions.DirectoryID.EQ(jet.UUID(n.CurrentUser.DirectoryID))),
		).
		RETURNING(table.Abstractions.ID, table.Abstractions.LastUpdatedOn)

	if err := updateQuery.Query(db, &n.AbstractionUpdate.Abstraction); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Update abstraction %v by %v failed | reason: %v", n.AbstractionID, n.CurrentUser.DirectoryID, err))
		return lib.NewError(http.StatusInternalServerError, "Could not update abstraction")
	}

	return nil
}

func (n *abstractions) createAbstractions() (int64, error) {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	editorUserRole := model.DirectoryProjectsRoles{}
	selectQuery := table.DirectoryProjectsRoles.
		SELECT(table.DirectoryProjectsRoles.CreatedOn).
		WHERE(
			table.DirectoryProjectsRoles.DirectoryID.EQ(jet.UUID(n.AbstractionsCreation.DirectoryID)).
				AND(table.DirectoryProjectsRoles.ProjectID.EQ(jet.UUID(n.AbstractionsCreation.ProjectID))).
				AND(table.DirectoryProjectsRoles.ProjectRoleID.EQ(jet.String(lib.ROLE_EDITOR))),
		)
	if err := selectQuery.Query(db, &editorUserRole); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get target user abstraction role in order to create abstractions from by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not get target user abstraction role")
	}
	if editorUserRole.CreatedOn.IsZero() {
		return -1, lib.NewError(http.StatusBadRequest, "User is not an editor")
	}

	if n.AbstractionsCreation.FilesSearchQuery != "" {
		n.AbstractionsCreation.Files = []model.Files{}
		whereClause := jet.BoolExp(lib.GetTextSearchBoolExp(table.Files.FileVector.Name(), n.AbstractionsCreation.FilesSearchQuery))
		if n.AbstractionsCreation.SkipFilesWithAbstractions {
			whereClause = whereClause.AND(table.Files.ID.NOT_IN(table.Abstractions.SELECT(table.Abstractions.FileID)))
		}
		selectQuery := table.Files.SELECT(table.Files.ID).WHERE(whereClause)
		if err := selectQuery.Query(db, &n.AbstractionsCreation.Files); err != nil {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get files to create abstractions from by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
			return -1, lib.NewError(http.StatusInternalServerError, "Could not get files to create abstractions from")
		}
	}
	if len(n.AbstractionsCreation.Files) < 1 {
		if n.AbstractionsCreation.SkipFilesWithAbstractions {
			return -1, lib.NewError(http.StatusBadRequest, "0 files found without abstractions")
		} else {
			return -1, lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}
	}
	abstractionsToCreate := []model.Abstractions{}
	for _, file := range n.AbstractionsCreation.Files {
		abstractionsToCreate = append(abstractionsToCreate, model.Abstractions{
			ProjectID:       n.AbstractionsCreation.ProjectID,
			DirectoryID:     n.AbstractionsCreation.DirectoryID,
			FileID:          file.ID,
			ModelTemplateID: n.AbstractionsCreation.ModelTemplateID,
			Tags:            &n.AbstractionsCreation.Tags,
			Abstraction:     "{}",
		})
	}

	insertQuery := table.Abstractions.
		INSERT(
			table.Abstractions.ProjectID,
			table.Abstractions.DirectoryID,
			table.Abstractions.FileID,
			table.Abstractions.ModelTemplateID,
			table.Abstractions.Tags,
			table.Abstractions.Abstraction,
		).MODELS(&abstractionsToCreate)

	if sqlResults, err := insertQuery.Exec(db); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Create abstractions by %v failed | reason: %v", n.CurrentUser.DirectoryID, err))
		return -1, lib.NewError(http.StatusInternalServerError, "Could not create abstractions")
	} else {
		if addedRows, err := sqlResults.RowsAffected(); err != nil {
			intpkglib.Log(intpkglib.LOG_WARNING, currentSection, fmt.Sprintf("Determining no. of abstractions created failed | reason: %v", err))
			return -1, nil
		} else {
			return addedRows, err
		}
	}
}
