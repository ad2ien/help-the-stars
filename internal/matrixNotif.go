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
	client       *mautrix.Client
	MatrixRoomID string
}

func CreateMatrixClient(ctx context.Context, settingsService *SettingsService) *MatrixClient {
	if !settingsService.GetSettings().IsMatrixConfigured() {
		return nil
	}

	client, err := mautrix.NewClient(settingsService.GetSettings().MatrixServer,
		id.UserID(settingsService.GetSettings().MatrixUsername),
		"")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Login(ctx, &mautrix.ReqLogin{
		Type: mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{
			User: settingsService.GetSettings().MatrixUsername,
			Type: mautrix.IdentifierTypeUser,
		},
		Password:         settingsService.GetSettings().MatrixPassword,
		StoreCredentials: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	client.DeviceID = resp.DeviceID

	return &MatrixClient{
		client:       client,
		MatrixRoomID: settingsService.GetSettings().MatrixRoomID,
	}
}

func (c *MatrixClient) Notify(ctx context.Context, issue *HelpWantedIssue) {
	if c == nil {
		log.Warn("Matrix not configured")

		return
	}

	message := fmt.Sprintf("Help wanted for new issue :\n%s\n%s", issue.Title, issue.Url)

	content := event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    message,
	}

	_, err := c.client.SendMessageEvent(ctx,
		id.RoomID(c.MatrixRoomID), event.EventMessage, content)
	if err != nil {
		log.Error(err)
	}
}

func (c *MatrixClient) NotifySeveralNewIssues(ctx context.Context) {
	if c == nil {
		log.Warn("Matrix not configured")

		return
	}

	message := "Help wanted for several new issues. Could be worth checking Help-the-stars interface ⭐"

	content := event.MessageEventContent{
		MsgType: event.MsgText,
		Body:    message,
	}

	_, err := c.client.SendMessageEvent(ctx,
		id.RoomID(c.MatrixRoomID), event.EventMessage, content)
	if err != nil {
		log.Error(err)
	}
}

func (c *MatrixClient) NotifyError(ctx context.Context, err error) {
	if c == nil {
		log.Warn("Matrix not configured")

		return
	}

	formattedMessage := fmt.Sprintf(`⚠️ Oups an error occurred:
	\n<blockquote>%s</blockquote>\n
	It will be retried next iteration...`, err)
	message := fmt.Sprintf("⚠️ Oups an error occurred: \n%s\nIt will be retried next iteration...", err)

	content := event.MessageEventContent{
		Format:        event.FormatHTML,
		FormattedBody: formattedMessage,
		MsgType:       event.MsgText,
		Body:          message,
	}

	_, err = c.client.SendMessageEvent(ctx,
		id.RoomID(c.MatrixRoomID), event.EventMessage, content)
	if err != nil {
		log.Error(err)
	}
}
