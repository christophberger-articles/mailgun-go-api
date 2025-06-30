package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// EmailMessage contains email headers and body parts
type EmailMessage struct {
	From           string
	To             []string
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
func (c *MailgunClient) SendRequest(ctx context.Context, msg MIMEMessage) (id string, err error) {
	req, err := http.NewRequest("POST", c.url, &msg.message)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", msg.contentType)
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

	// An ad-hoc struct for extracting the response ID
	var response struct {
		Id string
	}

	err = json.Unmarshal(responseBody, &response)

	return response.Id, nil

}
