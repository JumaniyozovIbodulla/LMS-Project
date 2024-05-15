package check

import (
	"backend_course/lms/config"
	"net/smtp"
)

func SendEmail(toEmail string, msg string) error {

	from := config.SmtpUsername
	to := []string{toEmail}
	subject := "Register for OTP"
	message := msg

	body := "To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" + message

	auth := smtp.PlainAuth("", config.SmtpUsername, config.SmtpPassword, config.SmtpServer)

	err := smtp.SendMail(config.SmtpServer+":"+config.SmtpPort, auth, from, to, []byte(body))
	if err != nil {
		return err
	}
	return nil
}
