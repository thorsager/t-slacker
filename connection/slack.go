package connection

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/config"
	"sort"
	"strings"
)

const cursorLimit = 1000

const (
	// TPublicChannel public channels
	TPublicChannel string = "public_channel"
	// TPrivateChannel private channels
	TPrivateChannel string = "private_channel"
	// TIm instant messages (privmsg) conversations
	TIm string = "im"
	// TMpim multi-point instant messages conversations
	TMpim string = "mpim"
)

var (
	// TAll all channels
	TAll = []string{TIm, TMpim, TPrivateChannel, TPublicChannel}
)

// Connection active connection to a slack-team
type Connection struct {
	Name              string
	authTest          *slack.AuthTestResponse
	User              *slack.User
	api               *slack.Client
	rtm               *slack.RTM
	onEvent           func(source *Connection, event *slack.RTMEvent)
	usersCache        []slack.User
	Config            *config.TeamConfig
	conversationCache map[string][]slack.Channel
}

// UserLookup lookup a team user by the userID
func (c *Connection) UserLookup(userID string) (*slack.User, error) {
	if len(c.usersCache) < 1 {
		var err error
		c.usersCache, err = c.GetUsers()
		if err != nil {
			return nil, err
		}
	}
	for _, e := range c.usersCache {
		if e.ID == userID {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", userID)
}

func (c *Connection) refreshUserCache() error {
	if len(c.usersCache) < 1 {
		var err error
		c.usersCache, err = c.GetUsers()
		if err != nil {
			return err
		}
	}
	return nil
}

// UserLookupByName look up team user by user-name
func (c *Connection) UserLookupByName(name string) (*slack.User, error) {
	if err := c.refreshUserCache(); err != nil {
		return nil, err
	}
	for _, e := range c.usersCache {
		if strings.ToUpper(e.Name) == strings.ToUpper(name) {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", name)
}

// GetUsers return all team users
func (c *Connection) GetUsers() ([]slack.User, error) {
	return c.api.GetUsers()
}

func (c *Connection) FindUserNamesStartingWith(prefix string) ([]string, error) {
	if err := c.refreshUserCache(); err != nil {
		return nil, err
	}
	var results []string
	for _, e := range c.usersCache {
		if strings.HasPrefix(strings.ToUpper(e.Name), strings.ToUpper(prefix)) {
			results = append(results, e.Name)
		}
	}
	sort.Strings(results)
	return results, nil
}

func asKey(segments ...string) string {
	sort.Strings(segments)
	return strings.Join(segments, "-")
}
func (c *Connection) refreshConversationCache(types ...string) error {
	key := asKey(types...)
	if len(c.conversationCache[key]) < 1 {
		conversations, err := c.GetConversations(types...)
		if err != nil {
			return err
		}
		c.conversationCache[key] = conversations
	}
	return nil
}

func (c *Connection) FindChannelNamesStartingWith(prefix string) ([]string, error) {
	channels, err := c.GetConversationsCaching(TPrivateChannel, TPublicChannel)
	if err != nil {
		return nil, err
	}

	var results []string
	for _, c := range channels {
		if strings.HasPrefix(strings.ToUpper(c.Name), strings.ToUpper(prefix)) {
			results = append(results, c.Name)
		}
	}
	sort.Strings(results)
	return results, nil
}

func (c *Connection) GetConversationsCaching(types ...string) ([]slack.Channel, error) {
	if err := c.refreshConversationCache(types...); err != nil {
		return nil, err
	}
	return c.conversationCache[asKey(types...)], nil
}

// GetAllChannels return all team channels
func (c *Connection) GetAllChannels() ([]slack.Channel, error) {
	return c.api.GetChannels(true)
}

// GetConversationInfo get channel information
func (c *Connection) GetConversationInfo(id string) (*slack.Channel, error) {
	return c.api.GetConversationInfo(id, true)
}

// GetConversations  get conversations
func (c *Connection) GetConversations(types ...string) ([]slack.Channel, error) {
	conversations, nextCursor, err := c.api.GetConversations(
		&slack.GetConversationsParameters{
			ExcludeArchived: "true",
			Limit:           cursorLimit,
			Types:           types,
		},
	)
	if err != nil {
		return nil, err
	}

	for nextCursor != "" {
		var additional []slack.Channel
		additional, nextCursor, err = c.api.GetConversations(
			&slack.GetConversationsParameters{
				ExcludeArchived: "true",
				Cursor:          nextCursor,
				Limit:           cursorLimit,
				Types:           types,
			},
		)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, additional...)
	}
	return conversations, nil
}

// GetMessages get channel messages
func (c *Connection) GetMessages(channelID string, count int) ([]slack.Message, error) {

	// https://godoc.org/github.com/nlopes/slack#GetConversationHistoryParameters
	historyParams := slack.GetConversationHistoryParameters{
		ChannelID: channelID,
		Limit:     count,
		Inclusive: false,
	}

	history, err := c.api.GetConversationHistory(&historyParams)
	if err != nil {
		return nil, err
	}

	// Reverse the order of the messages, we want the newest in
	// the last place
	var messagesReversed []slack.Message
	for i := len(history.Messages) - 1; i >= 0; i-- {
		messagesReversed = append(messagesReversed, history.Messages[i])
	}

	return messagesReversed, nil
}

// SendMessage send message to a channel
func (c *Connection) SendMessage(channelID, message string) error {
	// https://godoc.org/github.com/nlopes/slack#Client.PostMessage
	_, _, err := c.api.PostMessage(channelID,
		slack.MsgOptionAsUser(true),
		slack.MsgOptionUsername(c.User.Name),
		slack.MsgOptionParse(true),
		slack.MsgOptionPostMessageParameters(slack.PostMessageParameters{LinkNames: 1}),
		slack.MsgOptionText(message, true),
	)
	if err != nil {
		return err
	}
	return nil
}

// New Create new team connection
func New(config *config.TeamConfig, oe func(source *Connection, event *slack.RTMEvent)) (*Connection, error) {
	c := &Connection{Config: config, Name: config.Name, onEvent: oe, usersCache: make([]slack.User, 0),
		conversationCache: make(map[string][]slack.Channel)}
	c.api = slack.New(config.Token)

	at, err := c.api.AuthTest()
	if err != nil {
		return nil, fmt.Errorf("unable to authenticate, check your token: %s", err)
	}
	c.authTest = at

	u, err := c.api.GetUserInfo(c.authTest.UserID)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user info: %s", err)
	}
	c.User = u
	c.usersCache, err = c.GetUsers() // trigger cache seed
	if err != nil {
		return nil, fmt.Errorf("unable to seed users cache: %s", err)
	}

	err = c.api.SetUserAsActive()
	if err != nil {
		return nil, fmt.Errorf("unable to set current user active: %s", err)
	}

	c.rtm = c.api.NewRTM()
	go c.rtm.ManageConnection()
	go c.delegateEvents()

	return c, nil
}

func (c *Connection) delegateEvents() {
	for {
		select {
		case event := <-c.rtm.IncomingEvents:
			c.onEvent(c, &event)
		}
	}
}
