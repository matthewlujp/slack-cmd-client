package slack

import (
	"fmt"
	"io"
	"net/http"
)

const (
	slackApiBaseURL = "https://slack.com/api"
)

func buildURL(endpoint string) string {
	return fmt.Sprintf("%s/%s", slackApiBaseURL, endpoint)
}

func buildRequest(requestMethod, targetOp, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(requestMethod, buildURL(targetOp), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	return req, nil
}
