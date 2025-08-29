package internal

import (
	"context"
	"fmt"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type MatrixClient struct {
	client    *mautrix.Client
	respLogin *mautrix.RespLogin
}

func CreateMatrixClient() *MatrixClient {

	if GetSetting("MATRIX_HOMESERVER") == "" ||
		GetSetting("MATRIX_USERNAME") == "" ||
		GetSetting("MATRIX_PASSWORD") == "" {
		return nil
	}

	client, err := mautrix.NewClient(GetSetting("MATRIX_HOMESERVER"),
		id.UserID(GetSetting("MATRIX_USERNAME")),
		"")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Login(context.Background(), &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			User: GetSetting("MATRIX_USERNAME"),
			Type: mautrix.IdentifierTypeUser,
		},
		Password:         GetSetting("MATRIX_PASSWORD"),
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
		id.RoomID(GetSetting("MATRIX_ROOMID")), event.EventMessage, content)
	if err != nil {
		log.Println(err)
	}
}
