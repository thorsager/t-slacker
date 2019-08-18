package runtime

import (
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
)

type privMsgCommand struct {
	source *pane.Pane
	args   []string
}

func (c *privMsgCommand) Execute(ctx *AppRuntime) {
	if len(c.args) != 1 {
		c.source.Log(constants.Name, "USAGE: /privmsg <user>")
		return
	}
	var team *connection.Connection
	if c.source == ctx.PaneController.GetStatusPane() {
		team = ctx.GetActiveTeam()
	} else {
		team = ctx.GetTeam(c.source.TeamId)
	}

	name := c.args[0]
	if name[0] == constants.UserIndicatorChar {
		name = name[1:]
	}
	user, err := team.UserLookupByName(name)
	if err != nil {
		c.source.Logf(team.Name, "%s", err)
		return
	}

	cl, err := team.GetConversations(connection.TIm)
	if err != nil {
		c.source.Log(team.Name, "unable to list privmsgs")
		return
	}
	if ch, found := findChannel(cl, byUserId(user.ID)); found {
		existing := ctx.PaneController.GetByChannelId(ch.ID)
		if existing == nil {
			ctx.AddPane(team.User.TeamID, ch.ID, true)
		} else {
			ctx.PaneController.SetActive(existing)
		}
	} else {
		c.source.Log(constants.Name, "TODO: Create new conversation.")
	}
}

func byUserId(id string) func(c slack.Channel) bool {
	return func(e slack.Channel) bool {
		return e.User == id
	}
}
