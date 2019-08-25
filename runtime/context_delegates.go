package runtime

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/pane"
	"time"
)

func (c *AppRuntime) buildConsoleStatusLine(pane *pane.Pane) string {
	teamName := "*Not-Connected*"
	userName := "???"
	modeString := "??"
	if c.TeamCount() > 0 {
		at := c.GetActiveTeam()
		teamName = at.Name
		userName = at.User.Name
		modeString = userModeString(at.User)
	}
	return fmt.Sprintf(" [lightgreen][[-]%.2d:%.2d[lightgreen]][-] [lightgreen][[-]%s%s[lightgreen]][-] [lightgreen][[-]%d:%s (change with ^X)[libhtgreen]][-] [lightgreen][[-]Act: %s[lightgreen]][-]",
		time.Now().Hour(),
		time.Now().Minute(),
		userName,
		modeString,
		c.PaneController.GetActiveIndex()+1,
		teamName,
		c.PaneController.GetFormattedActivityString())
}

func (c *AppRuntime) buildStatusLine(pane *pane.Pane) string {
	at := c.GetTeam(pane.TeamId)
	teamName := "*Not-Connected*"
	userName := "???"
	modeString := "??"
	chanMod := "??"
	chanName := "???"
	if at != nil {
		teamName = at.Name
		userName = at.User.Name
		modeString = userModeString(at.User)
		ch, err := at.GetConversationInfo(pane.Channel.ID)
		if err == nil {
			chanName = ch.Name
			chanMod = channelMod(*ch)
		}
	}
	return fmt.Sprintf(" [lightgreen][[-]%.2d:%.2d[lightgreen]][-] [lightgreen][[-]%s%s[lightgreen]][-] [lightgreen][[-]%d:%s/%s(%s)[lightgreen]][-] [lightgreen][[-]Act: %s[lightgreen]][-]",
		time.Now().Hour(),
		time.Now().Minute(),
		userName,
		modeString,
		c.PaneController.GetActiveIndex()+1,
		teamName,
		chanName,
		chanMod,
		c.PaneController.GetFormattedActivityString())
}

func userModeString(u *slack.User) string {
	ms := ""
	if u.IsAdmin {
		ms += "A"
	}
	if u.IsOwner {
		ms += "o"
	}
	if u.IsPrimaryOwner {
		ms += "O"
	}
	if u.IsBot {
		ms += "b"
	}
	if ms != "" {
		return "([lightgreen]+[-]" + ms + ")"
	} else {
		return ""
	}
}
