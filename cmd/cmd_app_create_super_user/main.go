package main

import (
	"context"
	"log"

	intdoment "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/entities"
	intdomint "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/domain/interfaces"
	intcmdappcreatesuperuser "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/cmd_app_create_super_user"
	intrepopostgres "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/interfaces/repository/postgres"
	intlib "github.com/icipe-official/Data-Abstraction-Platform-Back-End/internal/lib"
)

func main() {
	ctx := context.TODO()

	logger := intlib.LogNewHttpLogger()

	repository, err := intrepopostgres.NewPostgresRepository(ctx, logger)
	if err != nil {
		log.Fatal("ERROR: Failed to establish repository connection, error: ", err)
	}

	var service intdomint.CreateSuperUserService = intcmdappcreatesuperuser.NewCmdCreateSuperUserService(repository, logger)

	var iamCredential *intdoment.IamCredentials
	for {
		if value, err := service.ServiceGetIamCredentials(ctx); err != nil {
			log.Fatal("ERROR: Get iam credential failed, error: ", err)
		} else if value == nil {
			log.Println("ERROR: iam credential not found")
		} else {
			iamCredential = value
			break
		}
	}

	successfulInserts, err := service.ServiceAssignSystemRolesToIamCredential(ctx, iamCredential)
	if err != nil {
		log.Fatal("ERROR: Assign system roles to  iam credential failed, error: ", err)
	}
	if err := service.ServiceCreateDirectoryForIamCredentials(ctx, iamCredential); err != nil {
		log.Println("Create Directory for Iam Credential failed, error:", err)
	}
	log.Printf("SUCCESS: %d system roles successfully assigned to iam credential with id: %v", successfulInserts, iamCredential.ID[0])
}
