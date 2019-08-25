package pane

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/nlopes/slack"
	"github.com/rivo/tview"
	"time"
)

type Pane struct {
	controller   *Controller
	layout       *tview.Flex
	content      *tview.TextView
	input        *tview.InputField
	title        *tview.TextView
	status       *tview.TextView
	name         string
	lastWrite    time.Time
	lastShow     time.Time
	notify       bool
	onInput      func(pane *Pane, input string)
	inputCapture func(pane *Pane, event *tcell.EventKey) *tcell.EventKey
	tabCapture   func(pane *Pane, event *tcell.EventKey) string
	statusLine   func(pane *Pane) string

	TeamId  string
	Channel slack.Channel
}

func (p *Pane) appendContent(s string) {
	_, _ = p.content.Write([]byte(s))
}

func newPane(ctrl *Controller, name, title string,
	sl func(p *Pane) string,
	oi func(p *Pane, i string),
	ic func(p *Pane, key *tcell.EventKey) *tcell.EventKey,
	tc func(pane *Pane, input string) string) *Pane {
	cp := &Pane{controller: ctrl, name: name, onInput: oi, inputCapture: ic, statusLine: sl}

	cp.content = tview.NewTextView().
		SetTextColor(tcell.ColorLightGray).
		SetScrollable(true).
		SetDynamicColors(true).
		SetWordWrap(true)

	cp.input = tview.NewInputField().
		SetLabel(fmt.Sprintf("[-][[-]%s[-]][-] ", name)).
		SetLabelColor(tcell.ColorLightGray).
		SetFieldTextColor(tcell.ColorLightGray).
		SetFieldBackgroundColor(tcell.ColorBlack)

	cp.input.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			text := cp.input.GetText()
			if len(text) > 0 {
				cp.onInput(cp, text)
				cp.input.SetText("")
			}
		}
	})

	cp.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyPgUp:
			or, _ := cp.content.GetScrollOffset()
			nor := or - 5
			cp.content.ScrollTo(nor, 0)
			return nil
		case tcell.KeyPgDn:
			or, _ := cp.content.GetScrollOffset()
			nor := or + 5
			cp.content.ScrollTo(nor, 0)
			return nil
		case tcell.KeyTAB:
			cp.input.SetText(tc(cp, cp.input.GetText()))
			return nil
		default:
			return cp.inputCapture(cp, event)
		}
	})

	cp.title = tview.NewTextView().SetTextColor(tcell.ColorLightGray)
	cp.title.SetBackgroundColor(tcell.ColorDarkGreen)
	cp.title.SetText(title)

	cp.status = tview.NewTextView().SetTextColor(tcell.ColorLightGray).SetDynamicColors(true)
	cp.status.SetBackgroundColor(tcell.ColorDarkGreen)
	cp.UpdateStatus()

	cp.layout = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cp.title, 1, 1, false).
		AddItem(cp.content, 0, 1, false).
		AddItem(cp.status, 1, 1, false).
		AddItem(cp.input, 1, 1, true)
	return cp
}
func h() int {
	return time.Now().Hour()
}
func m() int {
	return time.Now().Minute()
}

func (p *Pane) Changed() bool {
	return p.lastShow.Before(p.lastWrite)
}

func (p *Pane) Notify() bool {
	return p.notify
}
func (p *Pane) GetName() string {
	return p.name
}
func (p *Pane) UpdateStatus() {
	if p.statusLine != nil {
		p.status.SetText(p.statusLine(p))
	} else {
		p.status.SetText("statusLine func not set")
	}
}
func (p *Pane) RawLogf(format string, args ...interface{}) {
	p.RawLog(fmt.Sprintf(format, args...))

}
func (p *Pane) RawLog(message string) {
	_, _ = p.Write([]byte(fmt.Sprintf("\n%.2d:%.2d %s", h(), m(), message)))
}

func (p *Pane) Log(mark, message string) {
	_, _ = p.Write([]byte(fmt.Sprintf("\n%.2d:%.2d [#666666][[-]%s[#666666]][-] [blue]-[-]![blue]-[-] %s", h(), m(), mark, message)))
}

func (p *Pane) WriteNoChange(buffer []byte) (int, error) {
	return p.content.Write(buffer)
}

func (p *Pane) Write(buffer []byte) (int, error) {
	if p.controller.GetActive() != p {
		p.lastWrite = time.Now()
	}
	return p.content.Write(buffer)
}

func (p *Pane) ScrollToEnd() {
	p.content.ScrollToEnd()
}

func (p *Pane) SetNotify() {
	p.lastWrite = time.Now()
	p.notify = true
}

func (p *Pane) Logf(mark, format string, args ...interface{}) {
	p.Log(mark, fmt.Sprintf(format, args...))
}

func (p *Pane) shown() {
	p.lastShow = time.Now()
	p.notify = false
}
