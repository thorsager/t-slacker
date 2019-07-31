package runtime

type quitCommand struct{}

func (c *quitCommand) Execute(ctx *AppRuntime) {
	ctx.Stop()
}
