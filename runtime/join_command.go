package runtime

import (
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
)

type joinCommand struct {
	source *pane.Pane
	args   []string
}

func (c *joinCommand) Execute(ctx *AppRuntime) {
	if len(c.args) != 1 {
		ctx.PaneController.GetStatusPane().Log(constants.Name, "USAGE: /JOIN <channel>")
		return
	}
	if c.source == ctx.PaneController.GetStatusPane() {
		team := ctx.GetActiveTeam()
		cl, err := team.GetConversations(connection.TAll...)
		if err != nil {
			ctx.PaneController.GetActive().Log("ERROR", "unable to list conversations")
			return
		}
		ch, err := channelByName(c.args[0], cl)
		if err != nil {
			ctx.PaneController.GetActive().Logf("ERROR", "%s", err)
			return
		}
		existing := ctx.PaneController.GetByChannelId(ch.ID)
		if existing == nil {
			ctx.AddPane(team.User.TeamID, ch.ID, true)
		} else {
			ctx.PaneController.SetActive(existing)
		}
	} else {
		ctx.PaneController.GetActive().Log(constants.Name, "join command only supported on console pane")
	}
}
