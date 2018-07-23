package slack

const (
	// SlackAPIBaseURL is endpoint of Slack web api
	SlackAPIBaseURL = "https://slack.com/api"
)

// Option type object used to modify Client
type Option func(*Client) error

// BaseURL returns an option which sets base url in a Client object
func BaseURL(baseURL string) Option {
	return func(c *Client) error {
		c.baseURL = baseURL
		return nil
	}
}
