package slack_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/matthewlujp/slack-cmd-client/src/slack"
)

const (
	validToken        = "xoxo-valid-token1"
	validNoScopeToken = "xoxo-valid-noscope-token2"
	invalidToken      = "xoxo-invalid-token"
)

var (
	mux          *http.ServeMux
	server       *httptest.Server
	serverLogger = log.New(os.Stdout, "[server]", log.LstdFlags)
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	return func() {
		server.Close()
	}
}

func TestNewClient(t *testing.T) {
	// raise error on empty token
	if _, err := slack.NewClient("", nil); err == nil {
		t.Error("no error is raised when an empty token is provied")
	}
}

func extractToken(r *http.Request) string {
	v := r.Header.Get("Authorization")
	vs := strings.Split(v, " ")
	if len(vs) == 0 {
		return ""
	}
	return vs[len(vs)-1]
}

func checkRequestFormat(method, contentType string, fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			serverLogger.Printf("request method should be %s, got %s", method, r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Header.Get("Content-Type") != contentType {
			serverLogger.Printf("request content type should be %s, got %s", contentType, r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fn(w, r)
	}
}

func authenticate(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// token check
		if extractToken(r) == validNoScopeToken { // token with no adequate scope
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "missing_scope"})
			return
		} else if extractToken(r) != validToken { // invalid token
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "invalid_auth"})
			return
		}

		fn(w, r)
	}
}

func TestObtainWorkspaceInfo(t *testing.T) {
	teardown := setup()
	defer teardown()

	mux.HandleFunc("/team.info", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {

		type team struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Domain      string `json:"domain"`
			EmailDomain string `json:"email_domain"`
		}
		json.NewEncoder(w).Encode(&struct {
			Ok   bool `json:"ok"`
			Team team `json:"team"`
		}{
			Ok: true,
			Team: team{
				ID:          "1234",
				Name:        "team1",
				Domain:      "team1-hoge",
				EmailDomain: "",
			},
		})

	})))

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

	mux.HandleFunc("/users.list", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		type user struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			RealName string `json:"real_name"`
			IsBot    bool   `json:"is_bot"`
		}
		json.NewEncoder(w).Encode(&struct {
			Ok      bool   `json:"ok"`
			Members []user `json:"members"`
		}{
			Ok: true,
			Members: []user{
				user{ID: "USLACKBOT", Name: "slackbot", RealName: "slackbot", IsBot: true},
				user{ID: "1", Name: "taro", RealName: "yamada taro", IsBot: false},
			},
		})

	})))

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	expected := slack.Members{
		slack.User{ID: "USLACKBOT", Name: "slackbot", RealName: "slackbot", IsBot: true},
		slack.User{ID: "1", Name: "taro", RealName: "yamada taro", IsBot: false},
	}
	if members, err := client.GetMembers(); err != nil {
		t.Errorf("obtaining members failed, %s", err)
	} else if !reflect.DeepEqual(members, expected) {
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

	type user struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		RealName string `json:"real_name"`
		IsBot    bool   `json:"is_bot"`
	}

	// for providing user info
	mux.HandleFunc("/users.list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&struct {
			Ok      bool   `json:"ok"`
			Members []user `json:"members"`
		}{
			Ok: true,
			Members: []user{
				user{ID: "1", Name: "taro", RealName: "yamada taro", IsBot: false},
				user{ID: "2", Name: "jiro", RealName: "kayama jiro", IsBot: false},
				user{ID: "3", Name: "fumino", RealName: "kimura fumino", IsBot: false},
			},
		})
	})

	type purpose struct {
		Value string `json:"value"`
	}
	type channel struct {
		ID       string   `json:"id"`
		Name     string   `json:"name"`
		IsMember bool     `json:"is_member"`
		Members  []string `json:"members"`
		Purpose  purpose  `json:"purpose"`
	}
	mux.HandleFunc("/channels.list", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&struct {
			Ok       bool      `json:"ok"`
			Channels []channel `json:"channels"`
		}{
			Ok: true,
			Channels: []channel{
				channel{ID: "c1", Name: "channel1", IsMember: true, Members: []string{"1", "2", "3"}, Purpose: purpose{Value: "hoge 1"}},
				channel{ID: "c2", Name: "channel2", IsMember: false, Members: []string{"2", "3"}, Purpose: purpose{Value: "hoge 2"}},
			},
		})

	})))
	mux.HandleFunc("/conversations.list", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&struct {
			Ok       bool      `json:"ok"`
			Channels []channel `json:"channels"`
		}{
			Ok: true,
			Channels: []channel{
				channel{ID: "c3", Name: "channel3", IsMember: true, Members: []string{"1", "2"}, Purpose: purpose{Value: "hoge 3"}},
			},
		})

	})))
	mux.HandleFunc("/groups.list", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&struct {
			Ok     bool      `json:"ok"`
			Groups []channel `json:"groups"`
		}{
			Ok: true,
			Groups: []channel{
				channel{ID: "c4", Name: "channel4", IsMember: true, Members: []string{"1", "3"}, Purpose: purpose{Value: "hoge 4"}},
			},
		})

	})))
	mux.HandleFunc("/im.list", checkRequestFormat("GET", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		type im struct {
			ID   string `json:"id"`
			IsIM bool   `json:"is_im"`
			User string `json:"user"`
		}
		json.NewEncoder(w).Encode(&struct {
			Ok  bool `json:"ok"`
			Ims []im `json:"ims"`
		}{
			Ok: true,
			Ims: []im{
				im{ID: "c5", IsIM: true, User: "2"},
				im{ID: "c6", IsIM: true, User: "3"},
			},
		})

	})))

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

func compareChannelSlice(s1, s2 []slack.Channel) bool {
	if len(s1) != len(s2) {
		return false
	}

OuterLoop1:
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if reflect.DeepEqual(v1, v2) {
				continue OuterLoop1
			}
		}
		return false
	}

OuterLoop2:
	for _, v1 := range s2 {
		for _, v2 := range s1 {
			if reflect.DeepEqual(v1, v2) {
				continue OuterLoop2
			}
		}
		return false
	}

	return true
}

func TestSendMessage(t *testing.T) {
	teardown := setup()
	defer teardown()

	targetChannel := "c1"
	text := "hogefoobar"
	mux.HandleFunc("/chat.postMessage", checkRequestFormat("POST", "application/x-www-form-urlencoded", authenticate(func(w http.ResponseWriter, r *http.Request) {
		byteBody, _ := ioutil.ReadAll(r.Body)
		values, _ := url.ParseQuery(string(byteBody))

		if values.Get("channel") != targetChannel {
			serverLogger.Printf("channel expected %s, got %s", targetChannel, values.Get("channel"))
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "channel_not_found"})
			return
		}
		if values.Get("text") != text {
			serverLogger.Printf("text expected %s, got %s", text, values.Get("text"))
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}
		if values.Get("as_user") != "true" {
			serverLogger.Print("should be as a user, but not")
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}
		json.NewEncoder(w).Encode(&struct {
			Ok bool `json:"ok"`
		}{Ok: true})

	})))

	// valid token
	client, _ := slack.NewClient(validToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, text); err != nil {
		t.Errorf("sending message failed, %s", err)
	}

	// raise error on invalid token
	client, _ = slack.NewClient(invalidToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, text); err == nil {
		t.Errorf("no error raised on invalid token")
	}

	// raise error on no scope
	client, _ = slack.NewClient(validNoScopeToken, nil, slack.BaseURL(server.URL))
	if err := client.SendMessage(targetChannel, text); err == nil {
		t.Errorf("no error raised on inadequate scope")
	}

}

func TestUpload(t *testing.T) {
	teardown := setup()
	defer teardown()

	targetChannel := "c1"
	targetTitle := "titel1"
	filepath := "./client.go"
	targetInitialComment := "hoge"
	mux.HandleFunc("/files.upload", func(w http.ResponseWriter, r *http.Request) {
		// check token
		if token := r.FormValue("token"); token != validToken {
			serverLogger.Printf("token expected %s, got %s", validToken, token)
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "invalid_auth"})
			return
		}

		// check channel
		if channel := r.FormValue("channels"); channel != targetChannel {
			serverLogger.Printf("channel expected %s, got %s", targetChannel, channel)
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}

		// check title
		if title := r.FormValue("title"); title != targetTitle {
			serverLogger.Printf("title expected %s, got %s", targetTitle, title)
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}

		// check initial comment
		if initComment := r.FormValue("initial_comment"); initComment != targetInitialComment {
			serverLogger.Printf("initial comment expected %s, got %s", targetInitialComment, initComment)
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}

		// check file content
		uploadedFile, _, _ := r.FormFile("file")
		defer uploadedFile.Close()
		uploadedFileByteBody, _ := ioutil.ReadAll(uploadedFile)
		f, _ := os.Open("./client.go")
		defer f.Close()
		fileByteBody, _ := ioutil.ReadAll(f)
		if !bytes.Equal(uploadedFileByteBody, fileByteBody) {
			serverLogger.Printf("uploaded file %s does match not", filepath)
			json.NewEncoder(w).Encode(&struct {
				Ok    bool   `json:"ok"`
				Error string `json:"error"`
			}{Ok: false, Error: "fatal_error"})
			return
		}

		json.NewEncoder(w).Encode(&struct {
			Ok bool `json:"ok"`
		}{Ok: true})
	})

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
