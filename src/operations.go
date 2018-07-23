package main

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/matthewlujp/slack-cmd-client/src/slack"
)

// Add given token to token file and enable user to send message and upload file to a workspace.
// Token file is created in the home directory.
// If a token file does not exist in the home directory, a new file is created.
func registerToken(token string) error {
	c, err := slack.NewClient(token, nil, logger)
	if err != nil {
		return err
	}

	workspace, err := c.ObtainWorkspaceInfo()
	if err != nil {
		return err
	}

	conf := &config{}
	if err := loadConfig(conf); err != nil {
		logger.Printf("[registerToken] load config failed, %s", err)
		return err
	}
	// check whether tha workspace has already been registered
	registeredID := -1
	for i := range conf.Workspaces {
		if conf.Workspaces[i].ID == workspace.ID {
			registeredID = i
		}
	}
	if registeredID > -1 { // has been registered
		fmt.Printf("Are you sure to overwrite workspace %s ?  y/n ", workspace.Name)
		var ans string
		fmt.Scan(&ans)
		if ans != "y" {
			fmt.Println("Operation cancelled.")
			return nil
		}
		conf.Workspaces[registeredID].Name = workspace.Name
		conf.Workspaces[registeredID].Domain = workspace.Domain
		conf.Workspaces[registeredID].Token = workspace.Token
	} else { // is not registered
		fmt.Printf("Are you sure to add workspace %s ?  y/n ", workspace.Name)
		var ans string
		fmt.Scan(&ans)
		if ans != "y" {
			fmt.Println("Operation cancelled.")
			return nil
		}
		conf.Workspaces = append(conf.Workspaces, *workspace)
	}

	// set current context workspace if not set
	if conf.CurrentWorkspaceToken == "" {
		conf.CurrentWorkspaceToken = token
	}

	if err := saveConfig(conf); err != nil {
		logger.Printf("[registerToken] saving config failed, %s", err)
		return err
	}

	fmt.Printf("Workspace %s registered.", workspace.Name)
	return nil
}

// switchWorkspace shows a list of registered workspaces and let user choose one.
func switchWorkspace() error {
	// list registered workspaces
	conf := &config{}
	if err := loadConfig(conf); err != nil {
		logger.Printf("[switchWorkspace] loading config failed, %s", err)
		return err
	}
	if len(conf.Workspaces) < 1 {
		fmt.Println("No workspace is registered.")
		return nil
	}
	fmt.Println(conf)

	workspaceNames := make([]string, 0, len(conf.Workspaces))
	var currentWorkspaceName string
	for _, w := range conf.Workspaces {
		workspaceNames = append(workspaceNames, w.Name)
		if w.Token == conf.CurrentWorkspaceToken {
			currentWorkspaceName = w.Name
		}
	}
	prompt := promptui.Select{
		Label: fmt.Sprintf("Current workspace is \"%s\". Select Workspace", currentWorkspaceName),
		Items: workspaceNames,
	}
	selectedID, result, err := prompt.Run()
	if err != nil {
		logger.Printf("[switchWorkspace] selection went wrong")
	}

	// switch current workspace token
	conf.CurrentWorkspaceToken = conf.Workspaces[selectedID].Token
	fmt.Println(conf)
	if err := saveConfig(conf); err != nil {
		logger.Printf("[switchWorkspace] failed in saveing new context token")
		return err
	}

	fmt.Printf("Switched to %s", result)
	return nil
}

func listChannels() error {
	workspace, token, err := getCurrentWorkspace()
	if err != nil {
		logger.Printf("[listChannels] getting current workspace name and token failed, %s", err)
		return err
	}

	c, err := slack.NewClient(token, nil, logger)
	if err != nil {
		logger.Printf("[listChannels] building new client failed, %s", err)
		return err
	}
	channels, err := c.CollectChannels()
	if err != nil {
		logger.Printf("[listChannels] collecting channels failed, %s", err)
		return err
	}

	fmt.Printf("Channels you join in workspace %s are,\n", workspace)
	for _, ch := range channels {
		var desc string
		if ch.IsDirectMessage {
			desc = fmt.Sprintf("Direct message to %s.", ch.Name)
		} else {
			desc = ch.Purpose.Value
		}
		fmt.Printf("%s:  %s\n", ch.Name, desc)
	}

	return nil
}

func sendMessage(channelIDOrName, message string) error {
	_, token, err := getCurrentWorkspace()
	if err != nil {
		logger.Printf("[sendMessage] retrieving toke from config file failed, %s", err)
		return err
	}
	c, err := slack.NewClient(token, nil, logger)
	if err != nil {
		logger.Printf("[sendMessage] building client failed, %s", err)
		return err
	}

	channelName, channelID, err := toChannelNameAndID(channelIDOrName, c)
	if err != nil {
		logger.Printf("[sendMessage] %s", err)
		return err
	}

	fmt.Printf("Sending message to %s\n", channelName)
	// send message
	if err := c.SendMessage(channelID, message); err != nil {
		logger.Printf("[sendMessage] send request failed, %s", err)
		return err
	}
	fmt.Println("Message successfully sent")
	return nil
}

func uploadFile(channelIDOrName, filepath, title, comment string) error {
	_, token, err := getCurrentWorkspace()
	if err != nil {
		logger.Printf("[uploadFile] retrieving toke from config file failed, %s", err)
		return err
	}
	c, err := slack.NewClient(token, nil, logger)
	if err != nil {
		logger.Printf("[uploadFile] building client failed, %s", err)
		return err
	}

	channelName, channelID, err := toChannelNameAndID(channelIDOrName, c)
	if err != nil {
		logger.Printf("[uploadFile] %s", err)
		return err
	}
	fmt.Printf("Uploading %s to %s\n", filepath, channelName)

	uploadOptions := make(map[string]string)
	if title != "" {
		uploadOptions["title"] = title
	}
	if comment != "" {
		uploadOptions["initial_comment"] = comment
	}
	if err := c.UploadFile(channelID, filepath, uploadOptions); err != nil {
		logger.Printf("[uploadFile] uploading failed, %s", err)
		return err
	}
	fmt.Println("File upload completed.")
	return nil
}

func toChannelNameAndID(channelIDOrName string, c *slack.Client) (string, string, error) {
	channels, err := c.CollectChannels()
	if err != nil {
		logger.Printf("[toChannelID] collecting channel failed, %s", err)
		return "", "", err
	}
	var targetChannelID string
	var targetChannelName string
	for _, ch := range channels {
		if ch.ID == channelIDOrName || ch.Name == channelIDOrName {
			targetChannelID = ch.ID
			targetChannelName = ch.Name
		}
	}
	if targetChannelID == "" {
		logger.Printf("[toChannelNameAndID] channel %s does not exist", channelIDOrName)
		return "", "", errors.New("invalid channel name or id")
	}
	return targetChannelName, targetChannelID, nil
}
