package connection

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/config"
	"os"
	"testing"
)

var (
	c *Connection
)

//func TestSlack(t *testing.T) {
//	token := os.Getenv("SLACK_TOKEN")
//	if token == "" {
//		t.Skip("Skipping testing, no SLACK_TOKEN")
//	}
//	client,err := New(token)
//	if err != nil {
//		t.Errorf("%+v",err)
//	}
//	fmt.Printf("%+v",client)
//}

func TestGetUsers(t *testing.T) {
	defer setup(t)(t)
	users, err := c.GetUsers()
	if err != nil {
		t.Errorf("%+v", err)
	}
	for i, u := range users {
		t.Logf("%d: %+v\n", i, u)
	}
}

func TestGetChannels(t *testing.T) {
	defer setup(t)(t)
	channels, err := c.GetConversations(TPublicChannel)
	if err != nil {
		t.Errorf("%+v", err)
	}
	for i, u := range channels {
		t.Logf("%d: %+v\n", i, u)
	}
}

func TestGetConversations(t *testing.T) {
	defer setup(t)(t)
	conversations, err := c.GetConversations(TAll...)
	if err != nil {
		t.Errorf("%+v", err)
	}
	for i, u := range conversations {
		t.Logf("%d: %+v\n", i, u)
	}
}

func setup(t *testing.T) func(t2 *testing.T) {
	t.Log("Setting up...")
	token := os.Getenv("SLACK_TOKEN")
	teamName := os.Getenv("SLACK_TEAM")
	if token == "" || teamName == "" {
		t.Skip("Skipping testing, no SLACK_TOKEN or SLACK_TEAM")
	}
	cl, err := New(&config.TeamConfig{
		Name:           teamName,
		Token:          token,
		AutoConnect:    true,
		AutoJoin:       false,
		Colorize:       false,
		ColorizeInline: false,
		History:        config.History{Fetch: false, Size: 0},
	}, func(s *Connection, e *slack.RTMEvent) {})
	if err != nil {
		t.Errorf("%+v", err)
	}
	c = cl

	return func(t *testing.T) {
		fmt.Printf("Tearing down..")
	}
}
