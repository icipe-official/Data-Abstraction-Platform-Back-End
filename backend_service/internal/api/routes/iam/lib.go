package iam

import "data_administration_platform/internal/pkg/data_administration_platform/public/model"

const currentSection string = "Identity and Access Management"

const (
	email_verification string = "email_verification"
	password_reset     string = "password_reset"
)

type iam struct {
	Login struct {
		Email    string
		Password string
	}
	IamRequestResponse struct {
		Email                string
		Password             string
		DirectoryIamTicketID string
		TicketNumber         string
		Pin                  string
	}
	RequestType string

	DirectoryIam model.DirectoryIam
}
