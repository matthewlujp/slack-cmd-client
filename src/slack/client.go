package slack

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Client is a wrapper for Slack web api
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
	logger     *log.Logger
}

// NewClient returns a client object to call Slack web api.
// A bearer token should be obtained here: https://api.slack.com/apps, and provide as an argument.
// You should grant several scopes as well.
// logger is a pointer to a logging object for debug, which can be nil if you don't need.
func NewClient(token string, logger *log.Logger, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, errors.New("invalid token")
	}
	if logger == nil {
		logger = log.New(ioutil.Discard, "", log.LstdFlags) // ignore error messages
	}
	c := &Client{
		token:      token,
		httpClient: http.DefaultClient,
		baseURL:    SlackAPIBaseURL,
		logger:     logger,
	}

	// parse options
	for _, option := range opts {
		if err := option(c); err != nil {
			logger.Printf("[NewClient] parsing option failed, %s", err)
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/%s", c.baseURL, endpoint)
}

func (c *Client) buildRequest(requestMethod, targetOp, token string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(requestMethod, c.buildURL(targetOp), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-type", "application/x-www-form-urlencoded")
	return req, nil
}

func (c *Client) get(method string) (*http.Response, error) {
	req, err := c.buildRequest("GET", method, c.token, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %s", res.Status)
	}
	return res, nil
}

func (c *Client) post(method string, body io.Reader) (*http.Response, error) {
	req, err := c.buildRequest("POST", method, c.token, body)
	if err != nil {
		return nil, err
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status %s", res.Status)
	}
	return res, nil
}

// ObtainWorkspaceInfo gets the current workspace info from Slack api.
// team:read scope should be granted beforehand.
// See https://api.slack.com/methods/team.info
func (c *Client) ObtainWorkspaceInfo() (*Workspace, error) {
	res, err := c.get("team.info")
	if err != nil {
		c.logger.Printf("[ObtainWorkspaceInfo] request failed, %s", err)
		return nil, err
	}
	defer res.Body.Close()

	// parse json
	wInfo := &struct {
		Ok    bool      `json:"ok"`
		Error string    `json:"error"`
		Team  Workspace `json:"team"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(wInfo); err != nil {
		c.logger.Printf("[ObtainWorkspaceInfo] failed in decoding response json, %s", err)
		return nil, err
	}

	if !wInfo.Ok {
		c.logger.Printf("[ObtainWorkspaceInfo] response does not contain workspace info, %s", wInfo.Error)
		return nil, errors.New(wInfo.Error)
	}
	wInfo.Team.Token = c.token
	return &wInfo.Team, nil
}

// GetMembers gets members of the current workspace from Slack api.
// users:read scope should be granted beforehand.
// See https://api.slack.com/methods/users.list
func (c *Client) GetMembers() (Members, error) {
	res, err := c.get("users.list")
	if err != nil {
		c.logger.Printf("[GetMembers] request failed, %s", err)
		return nil, err
	}
	defer res.Body.Close()

	// parse json response
	parsed := &struct {
		Ok      bool    `json:"ok"`
		Error   string  `json:"error"`
		Members Members `json:"members"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(parsed); err != nil {
		c.logger.Printf("[GetMembers] parsing json response failed")
		return nil, err
	}
	if !parsed.Ok {
		c.logger.Printf("[GetMembers] request rejected by Slack, %s", parsed.Error)
		return nil, errors.New(parsed.Error)
	}
	return parsed.Members, nil
}

// CollectChannels collects channels which a user joins in the current workspace.
// It collects channels from channels.list, conversaions.list, groups.list, and im.list (direct message).
// channels:read, groups:read, im:read, and mpim:read scopes should be granted.
// See https://api.slack.com/methods/channels.list, https://api.slack.com/methods/groups.list, https://api.slack.com/methods/conversations.list, and https://api.slack.com/methods/im.list
func (c *Client) CollectChannels() ([]Channel, error) {
	collectedChannels := make(map[string]Channel)
	for _, m := range []string{"channels.list", "conversations.list", "groups.list", "im.list"} {
		chans, err := c.getChannels(m)
		if err != nil {
			c.logger.Printf("[CollectChannels] inquiring channels from %s failed, %s", m, err)
			return nil, err
		}

		for _, c := range chans {
			// only add channels that a user joins
			if _, ok := collectedChannels[c.ID]; !ok && (c.IsMember || c.IsDirectMessage) {
				collectedChannels[c.ID] = c
			}
		}
	}

	// edit if direct message
	var channels []Channel
	members, err := c.GetMembers()
	if err != nil {
		c.logger.Printf("[CollectChannels] obtaining members failed, %s", err)
		return nil, err
	}
	for _, c := range collectedChannels {
		if c.IsDirectMessage {
			if companionName, err := members.ID2UserName(c.User); err == nil {
				c.Name = companionName // user companion name as a channel name
			} else {
				c.Name = "Direct Message to ???"
			}

		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (c *Client) getChannels(method string) ([]Channel, error) {
	res, err := c.get(method)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	parsed := &struct {
		Ok       bool      `json:"ok"`
		Error    string    `json:"error"`
		Channels []Channel `json:"channels"`
		Groups   []Channel `json:"groups"`
		IMS      []Channel `json:"ims"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(parsed); err != nil {
		return nil, err
	}
	if !parsed.Ok {
		return nil, fmt.Errorf("request rejected by Slack, %s", parsed.Error)
	}

	if len(parsed.Channels) > 0 {
		return parsed.Channels, nil
	} else if len(parsed.Groups) > 0 {
		return parsed.Groups, nil
	} else if len(parsed.IMS) > 0 {
		return parsed.IMS, nil
	}
	return []Channel{}, nil
}

// SendMessage sends a message to a designated channel.
// chat:write:user scope should be granted
// See https://api.slack.com/methods/chat.postMessage
func (c *Client) SendMessage(channelID, content string) error {
	v := url.Values{}
	v.Set("channel", channelID)
	v.Set("text", content)
	v.Set("as_user", "true")
	res, err := c.post("chat.postMessage", strings.NewReader(v.Encode()))
	if err != nil {
		c.logger.Printf("[SendMessage] post failed, %s", err)
		return err
	}
	defer res.Body.Close()

	parsed := &struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(parsed); err != nil {
		c.logger.Printf("[SendMessage] decoding json response failed, %s", err)
		return err
	}
	if !parsed.Ok {
		c.logger.Printf("[SendMessage] file upload request rejected by Slack, %s", parsed.Error)
		return errors.New(parsed.Error)
	}

	return nil
}

// UploadFile uploads a file to a designated channel.
// files:write:user scope should be granted.
// See https://api.slack.com/methods/files.upload
func (c *Client) UploadFile(channelID, filepath string, uploadOptions map[string]string) error {
	bf := new(bytes.Buffer)
	multiWriter := multipart.NewWriter(bf)

	// write file
	fileWriter, err := multiWriter.CreateFormFile("file", filepath)
	if err != nil {
		c.logger.Printf("[UploadFile] creating form-data for file failed, %s", err)
		return err
	}
	f, err := os.Open(filepath)
	if err != nil {
		c.logger.Printf("[UploadFile] opening file failed, %s", err)
		return err
	}
	defer f.Close()
	if _, err := io.Copy(fileWriter, f); err != nil {
		c.logger.Printf("[UploadFile] copying file contents to form writer failed, %s", err)
		return err
	}

	// add upload options
	uploadOptions["channels"] = channelID
	uploadOptions["token"] = c.token
	for k, v := range uploadOptions {
		w, err := multiWriter.CreateFormField(k)
		if err != nil {
			c.logger.Printf("[UploadFile] making %s field failed, %s", k, err)
			return err
		}
		if _, err := io.Copy(w, strings.NewReader(v)); err != nil {
			c.logger.Printf("[UploadFile] writing option %s:%s failed, %s", k, v, err)
			return err
		}
	}

	contentType := multiWriter.FormDataContentType()
	multiWriter.Close()

	// post
	url := c.buildURL("files.upload")
	res, err := http.Post(url, contentType, bf)
	if err != nil {
		c.logger.Printf("[UploadFile] posting file failed, %s", err)
		return err
	}
	if res.StatusCode != http.StatusOK {
		c.logger.Printf("[UploadFile] response status %s", res.Status)
		return fmt.Errorf("file upload post status, %s", res.Status)
	}
	defer res.Body.Close()

	// parse and check respons
	parsed := &struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}{}
	if err := json.NewDecoder(res.Body).Decode(parsed); err != nil {
		c.logger.Printf("[UploadFile] decoding json response failed, %s", err)
		return err
	}
	if !parsed.Ok {
		c.logger.Printf("[UploadFile] file upload request rejected by Slack, %s", parsed.Error)
		return errors.New(parsed.Error)
	}
	return nil
}
