package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type MIMEMessage struct {
	message     bytes.Buffer
	contentType string
}

func (c *MailgunClient) SendEmail(ctx context.Context, msg EmailMessage) (id string, err error) {
	mime := MIMEMessage{}
	writer := multipart.NewWriter(&mime.message)

	writer.WriteField("from", msg.From)
	writer.WriteField("subject", msg.Subject)

	for _, recipient := range msg.To {
		writer.WriteField("to", recipient)
	}

	if msg.Text != "" {
		writer.WriteField("text", msg.Text)
	}
	if msg.HTML != "" {
		writer.WriteField("html", msg.HTML)
	}

	if msg.AttachmentPath != "" {
		file, err := os.Open(msg.AttachmentPath)
		if err != nil {
			return "", fmt.Errorf("failed to open attachment: %w", err)
		}
		defer file.Close()

		fileWriter, err := writer.CreateFormFile("attachment", file.Name())
		if err != nil {
			return "", fmt.Errorf("failed to create form file: %w", err)
		}

		if _, err := io.Copy(fileWriter, file); err != nil {
			return "", fmt.Errorf("failed to copy file: %w", err)
		}

	}

	mime.contentType = writer.FormDataContentType()

	writer.Close()

	id, err = c.SendRequest(ctx, mime)
	if err != nil {
		return "", err
	}

	return id, nil
}
