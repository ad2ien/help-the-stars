package internal

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type MatrixClient struct {
	client *mautrix.Client
}

func CreateMatrixClient() *MatrixClient {

	if !GetSettings().IsMatrixConfigured() {
		return nil
	}

	client, err := mautrix.NewClient(GetSettings().MatrixServer,
		id.UserID(GetSettings().MatrixUsername),
		"")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Login(context.Background(), &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			User: GetSettings().MatrixUsername,
			Type: mautrix.IdentifierTypeUser,
		},
		Password:         GetSettings().MatrixPassword,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	client.DeviceID = resp.DeviceID

	return &MatrixClient{
		client: client,
	}
}

func (c *MatrixClient) Notify(issue *HelpWantedIssue) {

	message := fmt.Sprintf("Help wanted for new issue :\n%s\n%s", issue.Title, issue.Url)

	content := event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    message,
	}
	_, err := c.client.SendMessageEvent(context.Background(),
		id.RoomID(GetSettings().MatrixRoomID), event.EventMessage, content)
	if err != nil {
		log.Error(err)
	}
}

func (c *MatrixClient) NotifySeveralNewIssues() {

	message := "Help wanted for several new issues. Could be worth checking Help-the-stars interface ⭐"

	content := event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    message,
	}
	_, err := c.client.SendMessageEvent(context.Background(),
		id.RoomID(GetSettings().MatrixRoomID), event.EventMessage, content)
	if err != nil {
		log.Error(err)
	}
}
