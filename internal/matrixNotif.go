package internal

import (
	"context"
	"log"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

type MatrixClient struct {
	client    *mautrix.Client
	respLogin *mautrix.RespLogin
}

func CreateMatrixClient() MatrixClient {

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

	return MatrixClient{
		client: client,
	}
}

func (c *MatrixClient) SendMsg(issue string) {
	content := event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    issue,
	}
	_, err := c.client.SendMessageEvent(context.Background(),
		id.RoomID(GetSetting("MATRIX_ROOMID")), event.EventMessage, content)
	if err != nil {
		log.Println(err)
	}
}
