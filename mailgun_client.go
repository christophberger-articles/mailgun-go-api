package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

// User data, required for bulk emails
type User struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

// Recipients holds the data of a recipient.
// By using a map type (rather than a struct with Email and User fields),
// function SendEmail() can trivially JSONify the Recipients into a
// Mailgun Recipient Variables data structure.
type Recipients map[string]User

// EmailMessage contains email headers and body parts
type EmailMessage struct {
	Sender         string
	Recipients     Recipients
	Subject        string
	Text           string
	HTML           string
	AttachmentPath string
}

// HTTPStatusError represents an HTTP status code and the related
// error message returned from failing requests.
type HTTPStatusError struct {
	StatusCode int
	Message    string
}

// Error() returns the string representation of an HTTPStatusError
func (e HTTPStatusError) Error() string {
	return fmt.Sprintf("mailgun error %d: %s", e.StatusCode, e.Message)
}

// MailgunClient is an http.Client with additional Mailgun settings
// (baseURL, domain, apiKey)
type MailgunClient struct {
	url    string
	apiKey string
	client *http.Client
}

// NewMailgunClient returns a new MailgunClient with a request
// timeout of 30 seconds.
func NewMailgunClient(baseURL, domain, key string) *MailgunClient {
	return &MailgunClient{
		url:    fmt.Sprintf("%s/v3/%s/%s", baseURL, domain, "messages"),
		apiKey: key,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendRequest sends an email
func (c *MailgunClient) SendRequest(message bytes.Buffer, contentType string) (id string, err error) {
	req, err := http.NewRequest("POST", c.url, &message)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth("api", c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", HTTPStatusError{
			StatusCode: resp.StatusCode,
			Message:    string(responseBody),
		}
	}

	// An ad hoc struct for extracting the response ID
	var response struct {
		Id string `json:"id"`
	}

	err = json.Unmarshal(responseBody, &response)
	return response.Id, nil
}

func (c *MailgunClient) SendEmail(msg EmailMessage) (id string, err error) {
	var message bytes.Buffer

	writer := multipart.NewWriter(&message)

	writer.WriteField("from", msg.Sender)
	writer.WriteField("subject", msg.Subject)

	for email, _ := range msg.Recipients {
		writer.WriteField("to", email)
	}
	if len(msg.Recipients) > 1 {
		// bulk email mode, use batch sending
		// create the Recipients Variable JSON
		rv, err := json.Marshal(msg.Recipients)
		if err != nil {
			return "", fmt.Errorf("SendEmail: %v", err)
		}
		writer.WriteField("recipient-variables", string(rv))
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

	contentType := writer.FormDataContentType()

	writer.Close()

	id, err = c.SendRequest(message, contentType)
	if err != nil {
		return "", err
	}

	return id, nil
}
