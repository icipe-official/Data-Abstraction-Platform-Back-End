package iam

import (
	"data_administration_platform/internal/api/lib"
	"data_administration_platform/internal/pkg/data_administration_platform/public/table"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"fmt"
	"net/http"

	jet "github.com/go-jet/jet/v2/postgres"
)

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
