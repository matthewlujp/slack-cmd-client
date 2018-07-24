package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/matthewlujp/slack-cmd-client/src/slack"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	configString = `CurrentWorkspaceToken = "xoxo-hoge-a"

[[workspaces]]
  ID = "000a"
  Name = "workspace A"
  Domain = "foo-bar"
  Token = "xoxo-hoge-a"
`
)

func setup() func() {
	// create tmp config file
	homeDir, _ := homedir.Dir()
	f, err := ioutil.TempFile(homeDir, "")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	configFilePath := f.Name()
	defaultTokenFile = filepath.Base(f.Name()) // overwrite config file path

	return func() {
		os.Remove(configFilePath)
	}
}

func TestLoadConfig(t *testing.T) {
	teardown := setup()
	defer teardown()

	configFilePath, err := getConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_TRUNC, 0775)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(configString); err != nil {
		panic(err)
	}
	f.Close()

	v := new(config)
	if err := loadConfig(v); err != nil {
		t.Error(err)
	}
	expected := config{
		CurrentWorkspaceToken: "xoxo-hoge-a",
		Workspaces: []slack.Workspace{
			slack.Workspace{
				ID:     "000a",
				Name:   "workspace A",
				Domain: "foo-bar",
				Token:  "xoxo-hoge-a",
			},
		},
	}
	if !reflect.DeepEqual(*v, expected) {
		t.Errorf("config expected %v, got %v", expected, *v)
	}

}

func TestSaveConfig(t *testing.T) {
	teardown := setup()
	defer teardown()

	conf := &config{
		CurrentWorkspaceToken: "xoxo-hoge-a",
		Workspaces: []slack.Workspace{
			slack.Workspace{
				ID:     "000a",
				Name:   "workspace A",
				Domain: "foo-bar",
				Token:  "xoxo-hoge-a",
			},
		},
	}

	if err := saveConfig(conf); err != nil {
		t.Error(err)
	}

	configFilePath, err := getConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	f, _ := os.Open(configFilePath)
	data, _ := ioutil.ReadAll(f)
	if string(data) != configString {
		t.Errorf("config file should be\n%s\n\nbut actually got\n\n%s", configString, string(data))
	}

}

func TestListWorkspaces(t *testing.T) {
	teardown := setup()
	defer teardown()

	configFilePath, err := getConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(configFilePath, os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(`CurrentWorkspaceToken = "xoxo-hoge-a"

[[workspaces]]
	ID = "000a"
	Name = "workspace A"
	Domain = "foo-bar"
	Token = "xoxo-hoge-a"

[[workspaces]]
	ID = "000b"
	Name = "workspace B"
	Domain = "hoge-foo"
	Token = "xoxo-hoge-b"
`); err != nil {
		panic(err)
	}
	f.Close()

	if ws, err := listWorkspaces(); err != nil {
		t.Error(err)
	} else if len(ws) != 2 || !reflect.DeepEqual(ws, []string{"workspace A", "workspace B"}) && !reflect.DeepEqual(ws, []string{"workspace B", "workspace A"}) {
		t.Errorf("workspaces expected \"workspace A\" and \"workspace B\", got %v", ws)
	}

}

func TestGetCurrentWorkspace(t *testing.T) {
	teardown := setup()
	defer teardown()

	configFilePath, err := getConfigFilePath()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile(configFilePath, os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	if _, err := f.WriteString(configString); err != nil {
		panic(err)
	}
	f.Close()

	if name, token, err := getCurrentWorkspace(); err != nil {
		t.Error(err)
	} else if name != "workspace A" || token != "xoxo-hoge-a" {
		t.Errorf("expected workspace A and xoxo-hoge-a, got %s and %s", name, token)
	}

}
