package slack_test

import (
	"testing"

	"github.com/matthewlujp/slack-cmd-client/src/slack"
)

func TestID2UserName(t *testing.T) {
	members := slack.Members{
		slack.User{ID: "1", Name: "a"},
		slack.User{ID: "2", Name: "b"},
	}
	if name, err := members.ID2UserName("1"); err != nil {
		t.Error(err)
	} else if name != "a" {
		t.Errorf("Expected user name a, but got %s", name)
	}
}
