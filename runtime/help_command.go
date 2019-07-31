package runtime

import "github.com/thorsager/t-slacker/constants"

type helpCommand struct{}

func (c *helpCommand) Execute(ctx *AppRuntime) {
	p := ctx.PaneController.GetActive()
	p.RawLogf("%s commands:", constants.Name)
	p.RawLog("/CONNECT <TEAM>")
	p.RawLog("/HELP")
	p.RawLog("/JOIN <channel>")
	p.RawLog("/PART")
	p.RawLog("/WINDOW [NEXT|PREV|LIST|#]")
	p.RawLog("/QUIT [NEXT|PREV|LIST|#]")
	p.RawLog("")
}
