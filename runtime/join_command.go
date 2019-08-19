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
		c.source.Log(constants.Name, "USAGE: /JOIN <channel>")
		return
	}
	var team *connection.Connection
	if c.source == ctx.PaneController.GetStatusPane() {
		team = ctx.GetActiveTeam()
	} else {
		team = ctx.GetTeam(c.source.TeamId)
	}
	if team == nil {
		c.source.Log(constants.Name, "unable to determine team")
		return
	}
	cl, err := team.GetConversations(connection.TAll...)
	if err != nil {
		c.source.Log(team.Name, "unable to list conversations")
		return
	}
	name := c.args[0]
	if name[0] == constants.ChannelIndicatorChar {
		name = name[1:]
	}
	ch, err := channelByName(name, cl)
	if err != nil {
		c.source.Logf(team.Name, "%s", err)
		return
	}
	existing := ctx.PaneController.GetByChannelId(ch.ID)
	if existing == nil {
		ctx.AddPane(team.User.TeamID, ch.ID, true)
	} else {
		ctx.PaneController.SetActive(existing)
	}
}
