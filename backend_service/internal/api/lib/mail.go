package lib

import (
	"crypto/tls"
	intpkglib "data_administration_platform/internal/pkg/lib"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"os"
)

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("unknown from server")
		}
	}
	return nil, nil
}

func SendEmail(subject, body string, to []string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", os.Getenv("MAIL_HOST"), os.Getenv("MAIL_PORT")))
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, "Email", fmt.Sprintf("Failed to dial server | reason: %v", err))
		return NewError(http.StatusInternalServerError, "Could not send email")
	}

	c, err := smtp.NewClient(conn, os.Getenv("MAIL_HOST"))
	if err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, "Email", fmt.Sprintf("Failed to create new smtp client | reason: %v", err))
		return NewError(http.StatusInternalServerError, "Could not send email")
	}

	if err = c.StartTLS(&tls.Config{
		ServerName:         os.Getenv("MAIL_HOST"),
		InsecureSkipVerify: true,
	}); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, "Email", fmt.Sprintf("Failed to make tls config | reason: %v", err))
		return NewError(http.StatusInternalServerError, "Could not send email")
	}

	auth := LoginAuth(os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"))
	if err = c.Auth(auth); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, "Email", fmt.Sprintf("Failed to auth client | reason: %v", err))
		return NewError(http.StatusInternalServerError, "Could not send email")
	}

	if err = smtp.SendMail(
		fmt.Sprintf("%v:%v", os.Getenv("MAIL_HOST"), os.Getenv("MAIL_PORT")),
		auth,
		os.Getenv("MAIL_USERNAME"),
		to,
		[]byte(fmt.Sprintf(EmailFormat, subject, body)),
	); err != nil {
		intpkglib.Log(intpkglib.LOG_ERROR, "Email", fmt.Sprintf("Failed to send mail | reason: %v", err))
		return NewError(http.StatusInternalServerError, "Could not send email")
	}
	return nil
}
