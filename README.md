# slack-cmd-client
Slack client with which you can send messages and upload files from terminal.
It is implemented in Golang.

# Dependencies
- golang 1.10
- [dep](https://github.com/golang/dep)

# Install
## Build from source
After cloning this repository, execute the following in the root directory.

```
make install
```

## Use built binaries
Pre-built binaries for some environments are included in bins/ directory.
Choose one of them which suits your environment and add it to the PATH environment variable.
The binaries are for...
- darwin amd64
- window amd64
- linux amd64



## Integration with your Slack account
To enable this tool to send messages and upload files to your workspaces, you should get bearer token and regiester it.
Here is the steps to do it.
### 1. Get a bearer token
Go to https://api.slack.com/apps and press "Create New App".
App name can be anything, and select a workspace to which you want to send messages and upload files via this tool.
Note workspace can be added as many times as you want laters.

### 2. Grant scopes
After creating an app, go to ["OAuth & Permissions"](https://api.slack.com/apps/ABWRAF13R/oauth?) to grant scopes to access Slack api.
Choose following scopes from dropdown list in the "Scope" section.
- team:read
- users:read
- channels:read
- groups:read
- im:read
- mpim:read
- chat:write:user
- files:write:user

After selecting the scopes, press "Save Changes" and then press "Install App to Workspace".

### 3. Register the token to this tool
Execute following to register the token acquired to the tool

```
% slack-cli add-token ****-**********-*********

Create new token file under the home directory.
Are you sure to add workspace LipTalk ?  y/n y
Workspace LipTalk registered.
```

This will create .slack-cmd-cli.toml under your home directory and add the workspace information to it.
If you want to register another workspace, simply follow the same procedure.
The workspace is locked to the first one as a default.
You can change this by using switch commend.


# Usage
Following subcommands are available.

## add-token
Register a workspace which is associated with a given token.
```
% slack-cli add-token *********************
```

## switch
This will show a list of registered workspaces and you can choose one from them to switch workspace.
```
% slack-cli switch

Switch workspace.
Use the arrow keys to navigate: ↓ ↑ → ←
? Current workspace is "Workspace A". Select Workspace:
  ▸ Workspace A
    Workspace B
    Workspace C
```

## list
List channels which you join. You can send or upload file to these channels.
```
% slack-cli list

Channels you join in workspace Workspace A are,
matthewlujp:  Direct message to matthewlujp.
general:  This channel is for workspace-wide communication and announcements. All members are in this channel.
random:  A place for non-work-related flimflam, faffing, hodge-podge or jibber-jabber you'd prefer to keep out of more focused work-related channels.
taro:  Direct message to slackbot.
fumino:  Direct message to nishizawa.
```

## message
Send message to a designated channel.
```
% slack-cli message taro "how are you doing?"

Sending message to taro
Message successfully sent
```

## upload
Upload a file to a designated channel.
You can designate the title with -t option and initial comment with -m option.
```
% slack-cli upload taro ./jiro_flying_in_the_sky.jpg -t "Flying Jiro" -m "Very interesting picture!"

Uploading ./jiro_flying_in_the_sky.jpg to taro
File upload completed.
```

# Let's Play!
Open a terminal and send a message or upload a file to your friends using while loop.
```
while :
do
  slack-cli message jiro "Hello! You are my best friend."
  sleep 10s
done
```
