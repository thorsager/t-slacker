package runtime

import (
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
	"strconv"
	"strings"
)

type windowCommand struct {
	source *pane.Pane
	args   []string
}

func (c *windowCommand) Execute(ctx *AppRuntime) {
	if len(c.args) < 1 {
		ctx.PaneController.GetStatusPane().Log(constants.Name, "USAGE: /WINDOW [NEXT|PREV|LIST|#]")
		return
	}
	switch strings.ToUpper(c.args[0]) {
	case "NEXT", "N":
		ctx.PaneController.SetNextActive()
	case "PREV", "P":
		ctx.PaneController.SetPrevActive()
	case "LIST", "LS", "L":
		sp := ctx.PaneController.GetStatusPane()
		for idx, p := range ctx.PaneController.GetPanes() {
			if idx != 0 {
				sp.RawLogf("\t%d: %s/%s ", idx+1, ctx.GetTeam(p.TeamId).Name, p.GetName())
			} else {
				sp.RawLogf("\t%d: %s ", idx+1, p.GetName())
			}
		}
	default:
		if i, err := strconv.Atoi(c.args[0]); err == nil {
			if i > 0 && i <= ctx.PaneController.GetSize() {
				ctx.PaneController.SetIndexActive(i - 1)
			} else {
				ctx.PaneController.GetActive().Logf("window not found: %s", c.args[0])
			}
		} else {
			ctx.PaneController.GetActive().Logf("window not found: %s", c.args[0])
		}
	}
}
