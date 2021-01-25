package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zachmann/mytoken/internal/utils"
)

// infoCommand is a type for holding and handling the info command
type infoCommand struct {
	generalOptions
	// EventHistory historyCommand `command:"history" description:"List the event history for this token"`
	// SubTree      treeCommand    `command:"tree" description:"List the tree of subtokens for this token"`
}

// Execute implements the flags.Commander interface
func (ic *infoCommand) Execute(args []string) error {
	_, superToken := ic.Check()
	if !utils.IsJWT(superToken) {
		return fmt.Errorf("The token is not a JWT.")
	}
	payload := strings.Split(superToken, ".")[1]
	decodedPayload, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return err
	}
	var infoBuffer bytes.Buffer
	if err = json.Indent(&infoBuffer, decodedPayload, "", "  "); err != nil {
		return err
	}
	info := infoBuffer.String()
	fmt.Println(info)
	return nil
}
