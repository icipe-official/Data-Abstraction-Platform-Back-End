package iam

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
}
