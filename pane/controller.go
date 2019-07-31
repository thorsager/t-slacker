package pane

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strings"
	"sync"
)

type Controller struct {
	sync.Mutex
	root            *tview.Pages
	panes           []*Pane
	currentPane     int
	queueUpdateDraw func(func())
	statusUpdate    func()
	onInput         func(pane *Pane, input string)
	inputCapture    func(pane *Pane, event *tcell.EventKey) *tcell.EventKey
}

func NewController(root *tview.Pages, qudFunc func(func()), oip func(p *Pane, i string), ic func(p *Pane, e *tcell.EventKey) *tcell.EventKey) *Controller {
	return &Controller{root: root, queueUpdateDraw: qudFunc, onInput: oip, inputCapture: ic}
}

func (c *Controller) AddPane(name, title, teamId, channelId string, show bool, statusLine func(p *Pane) string) *Pane {
	pane := newPane(c, name, title, statusLine, c.onInput, c.inputCapture)
	pane.TeamId = teamId
	pane.ChannelId = channelId
	c.Lock()
	c.panes = append(c.panes, pane)
	if show {
		c.currentPane = len(c.panes) - 1
	}
	c.Unlock()
	pane.UpdateStatus()
	if show {
		c.root.AddAndSwitchToPage(pane.name, pane.layout, true)
		pane.shown()
	} else {
		c.root.AddPage(pane.name, pane.layout, true, false)
	}
	return pane
}

func (c *Controller) SetActive(pane *Pane) {
	if idx, found := c.getIndex(pane); found {
		c.SetIndexActive(idx)
	}
}

func (c *Controller) RemovePane(pane *Pane) {

	if c.IsActive(pane) {
		c.SetPrevActive()
	}
	c.Lock()

	c.root.RemovePage(pane.name)

	if i, ok := c.getIndex(pane); ok {
		c.panes = append(c.panes[:i], c.panes[i+1:]...)
		//c.panes[len(c.panes)-1] = nil // or the zero value of T
		//c.panes = c.panes[:len(c.panes)-1]
	}

	c.Unlock()
	pane.UpdateStatus()
}

func (c *Controller) getIndex(pane *Pane) (int, bool) {
	for i, p := range c.panes {
		if p == pane {
			return i, true
		}
	}
	return -1, false
}

func (c *Controller) GetActive() *Pane {
	return c.panes[c.currentPane]
}

func (c *Controller) IsActive(pane *Pane) bool {
	return pane == c.panes[c.currentPane]
}

func (c *Controller) GetStatusPane() *Pane {
	return c.panes[0]
}

func (c *Controller) UpdateStatusBar() {
	c.panes[c.currentPane].UpdateStatus()
}

func (c *Controller) SetIndexActive(idx int) {
	c.Lock()
	pane := c.panes[idx]
	c.root.SwitchToPage(pane.name)
	c.currentPane = idx
	c.Unlock()
	pane.shown()
	c.UpdateStatusBar()
}

func (c *Controller) SetNextActive() {
	n := (c.GetActiveIndex() + 1) % len(c.panes)
	c.SetIndexActive(n)
	c.queueUpdateDraw(func() {})
}

func (c *Controller) SetPrevActive() {
	var n int
	if c.GetActiveIndex() < 1 {
		n = c.GetSize() - 1
	} else {
		n = c.GetActiveIndex() - 1
	}
	c.SetIndexActive(n)
	c.queueUpdateDraw(func() {})
}

func (c *Controller) SetStatusActive() {
	c.SetIndexActive(0)
}

func (c *Controller) GetActiveIndex() int {
	return c.currentPane
}

func (c *Controller) GetSize() int {
	return len(c.panes)
}

func (c *Controller) GetFormattedActivityString() string {
	var ls []string
	for i, p := range c.panes {
		if p.Changed() {
			if p.Notify() {
				ls = append(ls, fmt.Sprintf("*%d", i+1))
			} else {
				ls = append(ls, fmt.Sprintf("%d", i+1))
			}
		}
	}
	return strings.Join(ls, ",")
}

func (c *Controller) GetByTeamId(teamId string) (ret []*Pane) {
	for _, e := range c.panes {
		if e.TeamId == teamId {
			ret = append(ret, e)
		}
	}
	return
}

func (c *Controller) GetByChannelId(channelId string) *Pane {
	for _, e := range c.panes {
		if e.ChannelId == channelId {
			return e
		}
	}
	return nil
}

func (c *Controller) GetWindowList() string {
	var ls []string
	for i, p := range c.panes {
		ls = append(ls, fmt.Sprintf("%d: %s (TEAM) - %s", i+1, p.name, p.title.GetText(true)))
	}
	return strings.Join(ls, "\n")
}

func (c *Controller) Write(buffer []byte, change bool) (int, error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	var count int
	for _, pane := range c.panes {
		var i int
		var err error
		if change {
			i, err = pane.Write(buffer)
		} else {
			i, err = pane.Write(buffer)
		}
		if err != nil {
			return -1, fmt.Errorf("whil writing to %s: %v", pane.name, err)
		}
		count += i
	}
	return count, nil
}

func (c *Controller) withPaneInput(p *Pane, input string) {

}
