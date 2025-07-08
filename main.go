package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {

	domain := os.Getenv("MAILGUN_DOMAIN")
	key := os.Getenv("MAILGUN_API_KEY")

	client := NewMailgunClient(
		"https://api.mailgun.net",
		domain,
		key,
	)

	sender := "Mailgun Sandbox <postmaster@" + domain + ">"
	recipient1 := os.Getenv("MAILGUN_TEST_RECIPIENT_1")
	recipient2 := os.Getenv("MAILGUN_TEST_RECIPIENT_2")

	msg := EmailMessage{
		Sender: sender,
		Recipients: Recipients{
			recipient1: User{},
		},
		Subject: "Hello from Mailgun!",
		Text:    "This is a test email sent via Mailgun API.",
		HTML:    "<html><head></head><body><h1>Test</h1><p>This is a test email sent via the Mailgun API.</p></body>",
	}

	// Demonstrate single email
	fmt.Println("Sending single email...")
	if id, err := client.SendEmail(msg); err != nil {
		handleError("single email", err)
	} else {
		fmt.Println("✓ Single email sent successfully, message ID =", id)
	}

	// Demonstrate bulk email
	fmt.Println("Sending bulk email...")
	msg = EmailMessage{
		Sender: sender,
		Recipients: Recipients{
			recipient1: User{
				Name: "Alice",
				Id:   "1",
			},
			recipient2: User{
				Name: "Bob",
				Id:   "2",
			},
		},
		Subject: "A Bulk Email from Mailgun!",
		Text:    "This is a bulk email sent via Mailgun API.",
		HTML:    "<html><head></head><body><h1>Test</h1><p>This is a bulk email sent via the Mailgun API.</p></body>",
	}

	msg.Text = "Hello from Mailgun!"
	msg.HTML = ""

	if id, err := client.SendEmail(msg); err != nil {
		handleError("bulk email", err)
	} else {
		fmt.Println("✓ Bulk email sent successfully, message ID =", id)
	}

	// Demonstrate email with attachment
	fmt.Println("Sending email with attachment...")
	msg = EmailMessage{
		Sender: sender,
		Recipients: Recipients{
			recipient1: User{},
		},
		Subject:        "Email with Attachment",
		Text:           "Please find the flyer attached.",
		HTML:           "<html><head></head><body><h1>Attachment Test</h1><p>Please find the flyer attached.</p></body>",
		AttachmentPath: "gopher.webp",
	}

	if id, err := client.SendEmail(msg); err != nil {
		handleError("email with attachment", err)
	} else {
		fmt.Println("✓ Email with attachment sent successfully, message ID =", id)
	}
}

func handleError(operation string, err error) {
	httpErr := HTTPStatusError{}
	if errors.As(err, &httpErr) {
		fmt.Printf("%s: sending failed: HTTP error %d: %s\n", operation, httpErr.StatusCode, httpErr.Message)
		return
	}
	fmt.Printf("Sending failed: %s\n", err)
}
