package twitch

import (
	"errors"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nextthang/lurkmode/internal/message"
)

type Client struct {
	client      *twitch.Client
	channels    []string
	messageChan chan<- message.Message
}

type messageConstraint interface {
	twitch.PrivateMessage | twitch.UserNoticeMessage
}

func makeMessageHandler[T messageConstraint](ch chan<- message.Message) func(T) {
	return func(msg T) {
		var i any = &msg
		parsedMessage := message.NewMessage(i.(twitch.Message))
		if parsedMessage == nil {
			return
		}
		select {
		case ch <- parsedMessage:
			// Do nothing, message sent successfully
		case <-time.After(10 * time.Second):
			// TODO: Ths indicates that the application is locked up. Maybe let's find a way to recover from this somehow?
			panic("message channel is full for 10 seconds, this seems like like an unrecoverable state")
		}
	}
}

func NewClient(messageChan chan<- message.Message, channels ...string) *Client {
	twitchClient := twitch.NewAnonymousClient()
	if twitchClient == nil {
		return nil
	}

	twitchClient.OnPrivateMessage(makeMessageHandler[twitch.PrivateMessage](messageChan))
	twitchClient.OnUserNoticeMessage(makeMessageHandler[twitch.UserNoticeMessage](messageChan))

	twitchClient.Join(channels...)

	return &Client{
		client:      twitchClient,
		channels:    channels,
		messageChan: messageChan,
	}
}

func (c *Client) Connect() error {
	defer close(c.messageChan)
	err := c.client.Connect()
	if err != nil && !errors.Is(err, twitch.ErrClientDisconnected) {
		return err
	}
	return nil
}

func (c *Client) Disconnect() error {
	return c.client.Disconnect()
}

func (c *Client) AddChannel(channel string) {
	c.channels = append(c.channels, channel)
	if c.client != nil {
		c.client.Join(channel)
	}
}
