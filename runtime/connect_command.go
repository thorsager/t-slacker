package runtime

import (
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
	"strings"
)

type connectCommand struct {
	source *pane.Pane
	args   []string
}

func (c *connectCommand) Execute(ctx *AppRuntime) {
	if len(c.args) < 1 {
		ctx.PaneController.GetStatusPane().Log(constants.Name, "USAGE: /CONNECT <TEAM>")
		return
	}

	switch strings.ToUpper(c.args[0]) {
	case "LIST", "LS":
		for i, team := range ctx.Cfg.Teams {
			ctx.PaneController.GetStatusPane().Logf(constants.Name, "%d: %s", i, team.Name)
		}
	default:
		team, err := ctx.Cfg.GetTeamConfig(c.args[0])
		if err != nil {
			ctx.PaneController.GetStatusPane().Logf("ERROR", "unable to connect: %v", err)
		} else {
			ctx.ConnectTeam(team.Name)
		}
	}
}
