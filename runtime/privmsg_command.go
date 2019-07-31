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
		ctx.PaneController.GetStatusPane().Log(constants.Name, "USAGE: /privmsg <user>")
		return
	}
	if c.source == ctx.PaneController.GetStatusPane() {
		team := ctx.GetActiveTeam()
		user, err := team.UserLookupByName(c.args[0])
		if err != nil {
			ctx.PaneController.GetActive().Logf("ERROR", "%s", err)
			return
		}

		cl, err := team.GetConversations(connection.TIm)
		if err != nil {
			ctx.PaneController.GetActive().Log("ERROR", "unable to list privmsgs")
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
			ctx.PaneController.GetActive().Log("ERROR", "TODO: Create new conversation.")
		}
	} else {
		ctx.PaneController.GetActive().Log(constants.Name, "privmsg command only supported on console pane")
	}
}

func byUserId(id string) func(c slack.Channel) bool {
	return func(e slack.Channel) bool {
		return e.User == id
	}
}
