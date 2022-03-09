package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strconv"

	mail "github.com/xhit/go-simple-mail/v2"
)

//go:embed templates
var emailTemplatesFS embed.FS

func initMailserver() (*mail.SMTPServer, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil, errors.New("must define SMTP_HOST")
	}
	port, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil || port == 0 {
		return nil, errors.New("must define SMTP_HOST")
	}

	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	return server, nil
}

func (app *application) SendMail(from, to, subject, tmpl string, data interface{}) error {

	templateToRender := fmt.Sprintf("templates/%s.html.gohtml", tmpl)

	t, err := template.New("email-html").ParseFS(emailTemplatesFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	formattedMessage := tpl.String()

	templateToRender = fmt.Sprintf("templates/%s.plain.tmpl", tmpl)
	t, err = template.New("email-plain").ParseFS(emailTemplatesFS, templateToRender)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		app.errorLog.Println(err)
		return err
	}

	plainMessage := tpl.String()

	client, err := app.mailServer.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).
		AddTo(to).
		SetSubject(subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if email.Error != nil {
		return email.Error
	}
	// Call Send and pass the client
	err = email.Send(client)
	if err != nil {
		return err
	}

	return nil
}
