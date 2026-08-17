package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type Sender struct {
	host      string
	port      int
	username  string
	password  string
	fromName  string
	fromEmail string
}

type VerificationEmailData struct {
	Username        string
	VerificationURL string
}

func NewSender(
	host string,
	port int,
	username string,
	password string,
	fromName string,
	fromEmail string,
) *Sender {
	return &Sender{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromName:  fromName,
		fromEmail: fromEmail,
	}
}

func (sender *Sender) SendVerificationEmail(
	toEmail string,
	username string,
	verificationURL string,
) error {
	subject := "Verify your Feeder account"

	htmlBody, err := buildVerificationEmailHTML(
		VerificationEmailData{
			Username:        username,
			VerificationURL: verificationURL,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"build verification email: %w",
			err,
		)
	}

	message := []byte(
		fmt.Sprintf(
			"From: %s <%s>\r\n"+
				"To: %s\r\n"+
				"Subject: %s\r\n"+
				"MIME-Version: 1.0\r\n"+
				"Content-Type: text/html; charset=UTF-8\r\n"+
				"\r\n"+
				"%s",
			sender.fromName,
			sender.fromEmail,
			toEmail,
			subject,
			htmlBody,
		),
	)

	address := fmt.Sprintf(
		"%s:%d",
		sender.host,
		sender.port,
	)

	var auth smtp.Auth

	if sender.username != "" {
		auth = smtp.PlainAuth(
			"",
			sender.username,
			sender.password,
			sender.host,
		)
	}

	err = smtp.SendMail(
		address,
		auth,
		sender.fromEmail,
		[]string{toEmail},
		message,
	)
	if err != nil {
		return fmt.Errorf(
			"send SMTP message: %w",
			err,
		)
	}

	return nil
}

func buildVerificationEmailHTML(
	data VerificationEmailData,
) (string, error) {
	const emailTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta
        name="viewport"
        content="width=device-width, initial-scale=1.0"
    >
    <title>Verify your email</title>
</head>

<body style="
    margin: 0;
    padding: 0;
    background-color: #0b0d12;
    font-family: Arial, sans-serif;
    color: #f8fafc;
">
    <table
        width="100%"
        cellpadding="0"
        cellspacing="0"
        role="presentation"
    >
        <tr>
            <td
                align="center"
                style="padding: 40px 16px;"
            >
                <table
                    width="100%"
                    cellpadding="0"
                    cellspacing="0"
                    role="presentation"
                    style="
                        max-width: 560px;
                        background-color: #111318;
                        border: 1px solid #2a2e39;
                        border-radius: 24px;
                        padding: 32px;
                    "
                >
                    <tr>
                        <td>
                            <h1 style="
                                margin: 0 0 16px;
                                font-size: 28px;
                                color: #ffffff;
                            ">
                                Welcome to Feeder
                            </h1>

                            <p style="
                                margin: 0 0 16px;
                                color: #b6bbc8;
                                line-height: 1.6;
                            ">
                                Hi {{.Username}},
                            </p>

                            <p style="
                                margin: 0 0 28px;
                                color: #b6bbc8;
                                line-height: 1.6;
                            ">
                                Verify your email address to activate
                                your account and start using Feeder.
                            </p>

                            <a
                                href="{{.VerificationURL}}"
                                style="
                                    display: inline-block;
                                    padding: 14px 24px;
                                    background-color: #7c5cfc;
                                    color: #ffffff;
                                    text-decoration: none;
                                    border-radius: 14px;
                                    font-weight: bold;
                                "
                            >
                                Verify email address
                            </a>

                            <p style="
                                margin: 28px 0 8px;
                                color: #808695;
                                font-size: 13px;
                                line-height: 1.5;
                            ">
                                If the button does not work, copy this
                                link into your browser:
                            </p>

                            <p style="
                                margin: 0;
                                color: #a99bff;
                                font-size: 13px;
                                word-break: break-all;
                            ">
                                {{.VerificationURL}}
                            </p>

                            <p style="
                                margin: 28px 0 0;
                                color: #666d7d;
                                font-size: 12px;
                            ">
                                You can ignore this email if you did
                                not create a Feeder account.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

	tmpl, err := template.New(
		"verification-email",
	).Parse(emailTemplate)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	if err := tmpl.Execute(
		&buffer,
		data,
	); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
