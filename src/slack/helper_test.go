// Utilities such as a dummy http server for test and slice comparison are defined here.
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

	"github.com/matthewlujp/slack-cmd-client/src/slack"
)

const (
	validToken           = "xoxo-valid-token1"
	validNoScopeToken    = "xoxo-valid-noscope-token2"
	invalidToken         = "xoxo-invalid-token"
	targetChannel        = "c1"
	targetText           = "hogefoobar"
	targetTitle          = "titel1"
	filepath             = "./client.go"
	targetInitialComment = "hoge"
)

var (
	mux          *http.ServeMux
	server       *httptest.Server
	serverLogger = log.New(os.Stdout, "[server]", log.LstdFlags)
)

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	setupHandlers(mux)

	return func() {
		server.Close()
	}
}

func compareMembers(s1, s2 slack.Members) bool {
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

func setupHandlers(mux *http.ServeMux) {
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
				user{ID: "2", Name: "jiro", RealName: "kayama jiro", IsBot: false},
				user{ID: "3", Name: "fumino", RealName: "kimura fumino", IsBot: false},
			},
		})

	})))

	type user struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		RealName string `json:"real_name"`
		IsBot    bool   `json:"is_bot"`
	}

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
		if values.Get("text") != targetText {
			serverLogger.Printf("text expected %s, got %s", targetText, values.Get("text"))
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
}
