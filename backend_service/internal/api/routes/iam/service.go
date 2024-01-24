package iam

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/model"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"
	"os"

	jet "github.com/go-jet/jet/v2/postgres"
)

func (n *iam) processResetRequest() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var updateColumn jet.Column
	var setColumn jet.ColumnAssigment
	switch n.RequestType {
	case password_reset:
		if len(n.IamRequestResponse.Password) < 6 {
			return lib.NewError(http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
		}
		updateColumn = table.DirectoryIam.Password
		setColumn = table.DirectoryIam.Password.SET(jet.String(n.IamRequestResponse.Password))
	case email_verification:
		updateColumn = table.DirectoryIam.Email
		setColumn = table.DirectoryIam.IsEmailVerified.SET(jet.Bool(true))
	}

	updateQuery := table.DirectoryIam.UPDATE(
		updateColumn,
		table.DirectoryIam.TicketNumber,
		table.DirectoryIam.Pin,
	).SET(
		setColumn,
		table.DirectoryIam.TicketNumber.SET(jet.StringExp(jet.NULL)),
		table.DirectoryIam.Pin.SET(jet.StringExp(jet.NULL)),
	).WHERE(
		table.DirectoryIam.TicketNumber.EQ(jet.String(n.IamRequestResponse.DirectoryIamTicketID)).
			AND(jet.BoolExp(jet.Raw(fmt.Sprintf("pin = crypt('%v', pin)", n.IamRequestResponse.Pin)))),
	).RETURNING(table.DirectoryIam.DirectoryID)

	n.DirectoryIam = model.DirectoryIam{}

	if err = updateQuery.Query(db, &n.DirectoryIam); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Could not respond to request with ticket and pin for %v with email %v | reason: %v", n.RequestType, n.IamRequestResponse.Email, err))
		return lib.NewError(http.StatusInternalServerError, fmt.Sprintf("Could not respond to %v request", n.RequestType))
	}

	return nil
}

func (n *iam) sendRequestEmail() error {
	var subject string
	var body string

	switch n.RequestType {
	case password_reset:
		subject = "Vector Atlas - Data Abstraction Platform | Password reset"
		body = fmt.Sprintf(
			`
			<p>Hello %v,</p>
			<p>Kindly click <a href="%v">here</a> to reset your password.</p>
			<p>Regards.</p>
			<br>
			<strong>NB. If you did not make this request, kindly notify your project/system administrator.</strong>
			`, "user", fmt.Sprintf("%v%v/iam/password_reset?t=%v&p=%v", os.Getenv("DOMAIN_URL"), os.Getenv("BASE_PATH"), *n.DirectoryIam.TicketNumber, n.IamRequestResponse.Pin),
		)
	case email_verification:
		subject = "Vector Atlas - Data Abstraction Platform | Email verification"
		body = fmt.Sprintf(
			`
			<p>Hello %v,</p>
			<p>Kindly click <a href="%v">here</a> to verify your email.</p>
			<p>Regards.</p>
			<br>
			<strong>NB. If you did not make this request, kindly notify the project / system administrator.</strong>
			`, "user", fmt.Sprintf("%v%v/iam/email_verification?t=%v&p=%v", os.Getenv("DOMAIN_URL"), os.Getenv("BASE_PATH"), *n.DirectoryIam.TicketNumber, n.IamRequestResponse.Pin),
		)
	}

	intpkglib.Log(intpkglib.LOG_DEBUG, currentSection, fmt.Sprintf("Request link for %v: %v", n.DirectoryIam.DirectoryID, fmt.Sprintf("%v%v/iam/%v?t=%v&p=%v", os.Getenv("DOMAIN_URL"), os.Getenv("BASE_PATH"), n.RequestType, *n.DirectoryIam.TicketNumber, n.IamRequestResponse.Pin)))

	if err := lib.SendEmail(subject, body, []string{*n.DirectoryIam.Email}); err != nil {
		return err
	}
	return nil
}

func (n *iam) getUserTicketAndPin() error {
	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return err
	}
	defer db.Close()

	var updateQuery jet.UpdateStatement
	switch n.RequestType {
	case password_reset:
		updateQuery = table.DirectoryIam.UPDATE(
			table.DirectoryIam.Password,
			table.DirectoryIam.TicketNumber,
			table.DirectoryIam.Pin,
		).
			SET(
				table.DirectoryIam.Password.SET(jet.StringExp(jet.NULL)),
				table.DirectoryIam.TicketNumber.SET(jet.StringExp(jet.Raw("gen_random_string()"))),
				table.DirectoryIam.Pin.SET(jet.String(n.IamRequestResponse.Pin)),
			).
			WHERE(
				table.DirectoryIam.Email.EQ(jet.String(n.IamRequestResponse.Email)).
					AND(table.DirectoryIam.IsEmailVerified.IS_TRUE()),
			).RETURNING(
			table.DirectoryIam.DirectoryID,
			table.DirectoryIam.Email,
			table.DirectoryIam.TicketNumber,
			table.DirectoryIam.Pin,
		)
	case email_verification:
		updateQuery = table.DirectoryIam.UPDATE(
			table.DirectoryIam.IsEmailVerified,
			table.DirectoryIam.TicketNumber,
			table.DirectoryIam.Pin,
		).
			SET(
				table.DirectoryIam.IsEmailVerified.SET(jet.Bool(false)),
				table.DirectoryIam.TicketNumber.SET(jet.StringExp(jet.Raw("gen_random_string()"))),
				table.DirectoryIam.Pin.SET(jet.String(n.IamRequestResponse.Pin)),
			).
			WHERE(table.DirectoryIam.Email.EQ(jet.String(n.IamRequestResponse.Email))).
			RETURNING(
				table.DirectoryIam.DirectoryID,
				table.DirectoryIam.Email,
				table.DirectoryIam.TicketNumber,
				table.DirectoryIam.Pin,
			)
	}

	n.DirectoryIam = model.DirectoryIam{}

	if err = updateQuery.Query(db, &n.DirectoryIam); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Could not create ticket and pin for %v with email %v | reason: %v", n.RequestType, n.IamRequestResponse.Email, err))
		return lib.NewError(http.StatusInternalServerError, fmt.Sprintf("Could not create %v request", n.RequestType))
	}

	return nil
}

func (n *iam) getUserByEmailandPassword() (*lib.User, error) {
	if len(n.Login.Email) < 7 || len(n.Login.Password) < 7 {
		return nil, lib.NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}

	db, err := intpkglib.OpenDbConnection()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	selectQuery := jet.SELECT(
		table.DirectoryIam.DirectoryID.AS("user.directory_id"),
		table.DirectoryIam.Email.AS("user.iam_email"),
		table.Directory.Name.AS("user.name"),
		table.Directory.Contacts.AS("user.contacts"),
		table.DirectorySystemUsers.CreatedOn.AS("user.system_user_created_on"),
		table.Projects.ID.AS("project.project_id"),
		table.Projects.Name.AS("project.project_name"),
		table.Projects.Description.AS("project.project_description"),
		table.Projects.CreatedOn.AS("project.project_created_on"),
		table.DirectoryProjectsRoles.ProjectRoleID.AS("project_roles.project_role_id"),
		table.DirectoryProjectsRoles.CreatedOn.AS("project_roles.project_role_created_on"),
	).FROM(
		table.DirectoryIam.INNER_JOIN(table.Directory, table.DirectoryIam.DirectoryID.EQ(table.Directory.ID)).
			LEFT_JOIN(table.DirectorySystemUsers, table.DirectoryIam.DirectoryID.EQ(table.DirectorySystemUsers.DirectoryID)).
			LEFT_JOIN(
				table.DirectoryProjectsRoles.INNER_JOIN(
					table.Projects,
					table.DirectoryProjectsRoles.ProjectID.EQ(table.Projects.ID),
				),
				table.DirectoryProjectsRoles.DirectoryID.EQ(table.DirectoryIam.DirectoryID),
			),
	).WHERE(
		table.DirectoryIam.Email.EQ(jet.String(n.Login.Email)).
			AND(jet.BoolExp(jet.Raw(fmt.Sprintf("password = crypt('%v', password)", n.Login.Password)))).
			AND(table.DirectoryIam.IsEmailVerified.IS_TRUE()).
			AND(table.DirectoryIam.DirectoryIamTicketID.IS_NULL()).
			AND(table.DirectoryIam.TicketNumber.IS_NULL()).
			AND(table.DirectoryIam.Pin.IS_NULL()),
	)

	user := new(lib.User)
	if err = selectQuery.Query(db, user); err != nil {
		if err.Error() == lib.POSTGRES_NOT_FOUND_ERROR {
			return nil, lib.NewError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		} else {
			intpkglib.Log(intpkglib.LOG_ERROR, currentSection, fmt.Sprintf("Get user by email and password failed | reason: %v", err))
			return nil, lib.NewError(http.StatusInternalServerError, "Could not authenticate user")
		}
	}

	return user, nil
}
