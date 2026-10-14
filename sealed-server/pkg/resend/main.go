package resend

import (
	"errors"
	"os"

	"github.com/resend/resend-go/v2"
)

func SendEmailRequest(to []string, subject string, html string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return errors.New("RESEND_API_KEY not set")
	}
	resendSupportEmail := os.Getenv("RESEND_SUPPORT_MAIL")
	resendReplyToEmail := os.Getenv("RESEND_REPLY_TO_MAIL")
	if resendReplyToEmail == "" || resendSupportEmail == "" {
		return errors.New("RESEND_SUPPORT_MAIL or/and RESEND_REPLY_TO_MAIL not set")
	}
	client := resend.NewClient(apiKey)
	params := &resend.SendEmailRequest{
		From:    resendSupportEmail,
		To:      to,
		Html:    html,
		Subject: subject,
		ReplyTo: resendReplyToEmail,
	}

	_, err := client.Emails.Send(params)
	return err
}
