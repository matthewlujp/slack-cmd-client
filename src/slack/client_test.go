package slack_test

import (
	"reflect"
	"testing"

	"github.com/matthewlujp/slack-cmd-client/src/slack"
)

func TestNewClient(t *testing.T) {
	// raise error on empty token
	if _, err := slack.NewClient("", nil); err == nil {
		t.Error("no error is raised when an empty token is provied")
	}
}

func TestObtainWorkspaceInfo(t *testing.T) {
	teardown := setup()
	defer teardown()

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	expected := &slack.Workspace{
		ID:     "1234",
		Name:   "team1",
		Domain: "team1-hoge",
		Token:  validToken,
	}
	if info, err := client.ObtainWorkspaceInfo(); err != nil {
		t.Errorf("obtaining workspace failed, %s", err)
	} else if !reflect.DeepEqual(info, expected) {
		t.Errorf("on valid token, expected %v, got %v", *expected, info)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if _, err := client.ObtainWorkspaceInfo(); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if _, err := client.ObtainWorkspaceInfo(); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}
}

func TestGetMembers(t *testing.T) {
	teardown := setup()
	defer teardown()

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	expected := slack.Members{
		slack.User{ID: "USLACKBOT", Name: "slackbot", RealName: "slackbot", IsBot: true},
		slack.User{ID: "1", Name: "taro", RealName: "yamada taro", IsBot: false},
		slack.User{ID: "2", Name: "jiro", RealName: "kayama jiro", IsBot: false},
		slack.User{ID: "3", Name: "fumino", RealName: "kimura fumino", IsBot: false},
	}
	if members, err := client.GetMembers(); err != nil {
		t.Errorf("obtaining members failed, %s", err)
	} else if !compareMembers(members, expected) {
		t.Errorf("on valid token, expected %v, got %v", expected, members)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if _, err := client.GetMembers(); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if _, err := client.GetMembers(); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}
}

func TestCollectChannels(t *testing.T) {
	teardown := setup()
	defer teardown()

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	// suppose client is user 1, then collected channels should be c1, c3, c4. c5, and c6.
	expected := []slack.Channel{
		slack.Channel{ID: "c1", Name: "channel1", Members: []string{"1", "2", "3"}, IsMember: true, Purpose: slack.Purpose{Value: "hoge 1"}},
		slack.Channel{ID: "c3", Name: "channel3", Members: []string{"1", "2"}, IsMember: true, Purpose: slack.Purpose{Value: "hoge 3"}},
		slack.Channel{ID: "c4", Name: "channel4", Members: []string{"1", "3"}, IsMember: true, Purpose: slack.Purpose{Value: "hoge 4"}},
		slack.Channel{ID: "c5", Name: "jiro", IsDirectMessage: true, User: "2"},
		slack.Channel{ID: "c6", Name: "fumino", IsDirectMessage: true, User: "3"},
	}
	if channels, err := client.CollectChannels(); err != nil {
		t.Errorf("collecting channels failed, %s", err)
		// } else if !reflect.DeepEqual(channels, expected) {
	} else if !compareChannelSlice(channels, expected) {
		t.Errorf("on valid token, expected %v, got %v", expected, channels)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if _, err := client.CollectChannels(); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if _, err := client.CollectChannels(); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}
}

func TestSendMessage(t *testing.T) {
	teardown := setup()
	defer teardown()

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, targetText); err != nil {
		t.Errorf("sending message failed, %s", err)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, targetText); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, targetText); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}
}

func TestUpload(t *testing.T) {
	teardown := setup()
	defer teardown()

	opts := map[string]string{
		"title":           targetTitle,
		"initial_comment": targetInitialComment,
	}

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	if err := client.UploadFile(targetChannel, filepath, opts); err != nil {
		t.Errorf("uploading failed, %s", err)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if err := client.UploadFile(targetChannel, filepath, opts); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	opts["token"] = validNoScopeToken
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if err := client.UploadFile(targetChannel, filepath, opts); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}
}
