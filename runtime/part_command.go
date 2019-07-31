package runtime

import (
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/pane"
	"strings"
)

type partCommand struct {
	source *pane.Pane
	args   []string
}

func (c *partCommand) Execute(ctx *AppRuntime) {
	types := []string{connection.TPrivateChannel, connection.TPublicChannel}
	if len(c.args) > 0 && strings.ToUpper(c.args[0]) == "ALL" {
		types = append(types, connection.TIm, connection.TMpim)
	}
	if c.source == ctx.PaneController.GetStatusPane() {
		ctx.PaneController.GetActive().Log("ERROR", "unable to close status-window, try /quit, if you really want to go.")
	} else {
		ctx.RemovePane(c.source)
	}
}
