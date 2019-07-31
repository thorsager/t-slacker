package runtime

import (
	"fmt"
	"time"
)

type TickHandler interface {
	OnTick(tick time.Time, tickSize time.Duration, ctx *AppRuntime)
}

type StatusPaneTimeUpdater struct {
	then time.Time
}

func (t *StatusPaneTimeUpdater) OnTick(tick time.Time, tickSize time.Duration, ctx *AppRuntime) {
	if tick.Minute() != t.then.Minute() {
		t.then = tick
		ctx.App.QueueUpdateDraw(func() {
			ctx.PaneController.UpdateStatusBar()
		})
	}
}

type DateChangeLogger struct {
	then time.Time
}

func (t *DateChangeLogger) OnTick(tick time.Time, tickSize time.Duration, ctx *AppRuntime) {
	if tick.Day() != t.then.Day() {
		t.then = tick
		_, err := ctx.PaneController.Write(
			[]byte(fmt.Sprintf("\n[gray]Day changed to %s[-]", tick.Format("02 January 2006"))),
			false)
		if err != nil {
			ctx.PaneController.GetStatusPane().Logf("ERROR", "error changing date: %v", err)
		}
	}
}
