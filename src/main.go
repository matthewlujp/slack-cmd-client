// Package to upload a file to a designated channel.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	logger          = log.New(os.Stdout, "", log.LstdFlags)
	uploadCmd       = flag.NewFlagSet("uplaod", flag.ExitOnError)
	uploadFileTitle = uploadCmd.String("t", "", "designate a title for the uploaded file")
	uploadComment   = uploadCmd.String("m", "", "add initial comments to the uploaded file")
)

const (
	cmdUsage = `  a) add-token token: create token file under the home directory
  b) switch: switch context workspace (from registered token)
  c) list: list channels to which you can upload a file
  d) message channel_id_or_name message_content: send message to a designated channel
  e) upload channel_id_or_name file_path [-t title] [-m comment]: upload a file`
)

// Call this script with one of following subcommands
// add-token token: create token file under the home directory and add a given token to the file
// switch: switch context workspace (from registered token)
// list: list channels to which you can upload a file
// message channel_id_or_name: upload a file
// upload channel_id_or_name file_path -t title -m comment: upload a file
func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Please provide valid subcommands.\n%s\n", cmdUsage)
		os.Exit(1)
	}
	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Printf("Following subcommands are available.\n%s\n", cmdUsage)
		os.Exit(0)
	}

	switch os.Args[1] {
	case "add-token":
		fmt.Println("Create new token file under the home directory.")
		if len(os.Args) < 3 || os.Args[2] == "" {
			fmt.Println("Usage: add-token token\nPlease provide a valid token.")
			os.Exit(1)
		}
		if err := registerToken(os.Args[2]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "switch":
		fmt.Println("Switching workspace.")
		if err := switchWorkspace(); err != nil {
			logger.Fatalf("failed to switch workspace, %s", err)
		}
	case "list":
		if err := listChannels(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "message":
		if len(os.Args) < 4 {
			fmt.Println("Usage: message channel_id_or_name message_content")
			os.Exit(1)
		}
		if err := sendMessage(os.Args[2], os.Args[3]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case "upload":
		if len(os.Args) < 4 {
			fmt.Println("Usage: upload channel_id_or_name filepath [-t title] [-m comment]")
			os.Exit(1)
		}
		uploadCmd.Parse(os.Args[4:])
		if err := uploadFile(os.Args[2], os.Args[3], *uploadFileTitle, *uploadComment); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		fmt.Println(uploadFileTitle, uploadComment)
		fmt.Printf("Subcommand %s is not supported.\n%s", os.Args[1], cmdUsage)
		os.Exit(1)
	}
}
