// Save workspace information and workspace context in toml file.
// Data format:
// 		[[workspaces]]
// 		name = "workspace A"
// 		token = "xoxo-hoge-a"
// 		domain = "domain a"
//
// 		[[workspaces]]
// 		name = "workspace B"
// 		token = "xoxo-hoge-b"
// 		domain = "domain b"
//
//		current_workspace_token = "xoxo-hoge-a"
//
// Switch workspace by modifying current_workspace_token
package main

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/matthewlujp/slack-cmd-client/src/slack"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	defaultTokenFile = ".slack_uploader.toml"
)

type config struct {
	Workspaces            []slack.Workspace `toml:"workspaces"`
	CurrentWorkspaceToken string            `toml:"current_workspace_token`
}

func getConfigFilePath() (string, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		logger.Printf("error when locating home dir, %s", err)
		return "", err
	}
	return filepath.Join(homeDir, defaultTokenFile), nil
}

func loadConfig(v *config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		logger.Printf("[loadConfig] failed in getting path, %s", err)
		return err
	}

	if _, err := os.Stat(configPath); err != nil {
		// no config file exist yet
		return nil
	}

	if _, err := toml.DecodeFile(configPath, v); err != nil {
		logger.Printf("[loadConfig] failed in decoding, %s", err)
		return err
	}
	return nil
}

// saveConfig saves given config struct as a toml file under the home directory.
// If config file does not exist, create a new one.
func saveConfig(conf *config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		logger.Printf("[saveConfig], %v", err)
		return err
	}

	var f *os.File
	if _, err := os.Stat(configPath); err != nil {
		f, err = os.Create(configPath)
		if err != nil {
			logger.Printf("[saveConfig], %v", err)
			return err
		}
	} else {
		f, err = os.OpenFile(configPath, os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			logger.Printf("[saveConfig], %v", err)
			return err
		}
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	if err := toml.NewEncoder(w).Encode(conf); err != nil {
		logger.Printf("[saveConfig], %v", err)
		return err
	}
	return nil
}

func listWorkspaces() ([]string, error) {
	conf := &config{}
	if err := loadConfig(conf); err != nil {
		return nil, err
	}

	var workspaces []string
	for _, ws := range conf.Workspaces {
		workspaces = append(workspaces, ws.Name)
	}
	return workspaces, nil
}

// getCurrentWorkspace returns current workspace name, its token, and an error if any
func getCurrentWorkspace() (string, string, error) {
	conf := &config{}
	if err := loadConfig(conf); err != nil {
		logger.Printf("[getCurrentWorkspaceToken] loading config failed, %s", err)
		return "", "", err
	}

	// search workspace name
	for _, w := range conf.Workspaces {
		if w.Token == conf.CurrentWorkspaceToken {
			return w.Name, conf.CurrentWorkspaceToken, nil
		}
	}
	return "", conf.CurrentWorkspaceToken, nil
}
