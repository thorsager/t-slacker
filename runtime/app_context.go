package runtime

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/nlopes/slack"
	"github.com/rivo/tview"
	"github.com/thorsager/t-slacker/config"
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
	"path"
	"sync"
	"time"
)

type AppRuntime struct {
	sync.Mutex
	AppHome        string
	Cfg            *config.Config
	App            *tview.Application
	Root           *tview.Pages
	PaneController *pane.Controller
	ticker         *time.Ticker
	tickHandlers   []TickHandler
	tickSize       time.Duration
	teams          []*connection.Connection
	currentTeam    int
	tab            tabControl
}
type tabControl struct {
	Token        string
	LastReturned string
	Count        int
	Indicator    uint8
}

// New Create new AppRuntime
func New(appHome string, title string, debug bool) (*AppRuntime, error) {
	cfg, err := config.Load(path.Join(appHome, "config.json"))
	if err != nil {
		return nil, err
	}
	cfg.Debug = debug || cfg.Debug

	app := tview.NewApplication()
	root := tview.NewPages()

	ctx := &AppRuntime{
		AppHome:  appHome,
		Cfg:      cfg,
		App:      app,
		Root:     root,
		tickSize: 100 * time.Millisecond,
		tab:      tabControl{Token: "", Count: 0},
	}

	controller := pane.NewController(root, ctx.queueUpdateDraw, ctx.onInput, ctx.inputCapture, ctx.tabCapture)
	ctx.PaneController = controller
	ctx.tickHandlers = append(ctx.tickHandlers, &StatusPaneTimeUpdater{then: time.Now()})
	ctx.tickHandlers = append(ctx.tickHandlers, &DateChangeLogger{then: time.Now()})
	ctx.PaneController.AddPane("(console)", title, constants.ConsoleTeamChannelId, nil, true, ctx.buildConsoleStatusLine)

	app.SetRoot(root, true)
	return ctx, nil
}

func (c *AppRuntime) AddPane(teamId, conversationId string, show bool) {
	t := c.GetTeam(teamId)
	tc, err := c.Cfg.GetTeamConfig(t.Name)
	if err != nil {
		c.PaneController.GetStatusPane().Logf("ERROR", "unable to find team config for %s: %v", teamId, err)
	}
	chnl, err := t.GetConversationInfo(conversationId)
	if err != nil {
		c.PaneController.GetStatusPane().Logf("ERROR", "unable to locate conversation %s on %s: %v", conversationId, teamId, err)
	}
	cname := constants.ChannelIndicator + chnl.Name
	topic := chnl.Topic.Value
	if chnl.IsIM {
		usr, err := t.UserLookup(chnl.User)
		if err != nil {
			cname = constants.UserIndicator + chnl.ID
			topic = fmt.Sprintf("Private message with \"<%s>\"", cname)
		} else {
			cname = constants.UserIndicator + usr.Name
			topic = fmt.Sprintf("Private message with \"%s <%s>\"", usr.RealName, cname)
		}
	}
	p := c.PaneController.AddPane(cname, topic, teamId, chnl, show, c.buildStatusLine)
	if tc.History.Fetch {
		messages, err := t.GetMessages(chnl.ID, tc.History.Size)
		if err != nil {
			c.PaneController.GetStatusPane().Logf("ERROR", "unable to fetch channel history for: %s", chnl.ID)
		} else {
			var lt time.Time
			for _, m := range messages {
				ct := asTime(m.Timestamp)
				if ct.Day() != lt.Day() {
					_, _ = p.WriteNoChange([]byte(fmt.Sprintf("\n[gray]Day changed to %s[-]", ct.Format("02 January 2006"))))
					lt = ct
				}
				if c.Cfg.Debug {
					c.PaneController.GetStatusPane().Logf("DEBUG", "history-fetch %s/#%s\n[#333333]%+v[-]", tc.Name, cname, m)
				}
				_, _ = p.WriteNoChange(c.renderMessage(t.User.TeamID, m))
			}
			now := time.Now()
			if lt.Year() != now.Year() || lt.Month() != now.Month() || lt.Day() != now.Day() {
				_, _ = p.WriteNoChange([]byte(fmt.Sprintf("\n[gray]Day changed to %s[-]", now.Format("02 January 2006"))))
			}
			p.ScrollToEnd()
		}
	}
	if c.Cfg.Debug {
		c.Debugf("new pane %+v", p)
	}
}

func (c *AppRuntime) RemovePane(p *pane.Pane) {
	c.PaneController.RemovePane(p)
	c.PaneController.GetStatusPane().Logf("DEBUG", "'%s' was removed", p.Channel.ID)
}

func channelByName(name string, l []slack.Channel) (slack.Channel, error) {
	for _, e := range l {
		if e.Name == name {
			return e, nil
		}
	}
	return *&slack.Channel{}, fmt.Errorf("channel %s not found in list", name)
}

func findChannel(l []slack.Channel, predicate func(channel slack.Channel) bool) (slack.Channel, bool) {
	for _, e := range l {
		if predicate(e) {
			return e, true
		}
	}
	return *&slack.Channel{}, false
}

func (c *AppRuntime) findTeamByUserID(userId string) *connection.Connection {
	for _, t := range c.teams {
		if t.User.ID == userId {
			return t
		}
	}
	return nil
}

// This is a "delegate" method to handle RTMEvents
func (c *AppRuntime) rtmEvent(source *connection.Connection, event *slack.RTMEvent) {
	if c.Cfg.Debug {
		c.PaneController.GetStatusPane().Logf(source.Name, "GOT EVENT: %s\n[#333333]%+v[-]", event.Type, event.Data)
	}
	switch event.Type {
	case "message":
		e := event.Data.(*slack.MessageEvent)
		if e.SubType == "message_replied" {
			c.PaneController.GetStatusPane().Logf("DEBUG", "%+v", e)
			c.PaneController.GetStatusPane().Logf("DEBUG", "%+v", e.SubMessage)
		} else {
			p := c.PaneController.GetByChannelId(e.Channel)
			if p == nil {
				cfg, err := c.Cfg.GetTeamConfig(source.Name)
				if err != nil {
					c.PaneController.GetStatusPane().Logf("ERROR", "unable to find pane or config from: %s", e.Channel)
					return
				}
				if cfg.AutoJoin {
					c.AddPane(source.User.TeamID, e.Channel, false)
					c.PaneController.GetStatusPane().Logf(source.Name, "opened chat for %s/#(%s)", source.Name, e.Channel)
				}
				return
			}
			_, _ = p.Write(c.renderMessageEvent(e))
			if c.Cfg.Notify {
				if found := c.findTeamByUserID(e.User); found == nil {
					fmt.Print("\a")
				}
			}
			if c.PaneController.IsActive(p) {
				c.queueUpdateDraw(func() {})
				team := c.GetTeam(p.TeamId)
				err := markConversation(team, p.Channel, e.Timestamp)
				if err != nil {
					c.PaneController.GetStatusPane().Logf("ERROR", "unable to set read mark for %s/%s ts:%s (%s)", p.TeamId, p.Channel.ID, e.Timestamp, err)
				} else {
					if c.Cfg.Debug {
						c.Debugf("set read marker for %s/%s ts:%s", p.TeamId, p.Channel.ID, e.Timestamp)
					}
				}
			}
		}
	default:
	}
	c.queueUpdateDraw(c.PaneController.UpdateStatusBar)
}

func markConversation(con *connection.Connection, cha slack.Channel, ts string) error {
	if cha.IsIM {
		return con.MarkIMChannel(cha.ID, ts)
	} else if cha.IsGroup {
		return con.SetGroupReadMark(cha.ID, ts)
	} else {
		return con.SetChannelReadMark(cha.ID, ts)
	}
}

// Connect to a team, and store connection i list of connected teams
func (c *AppRuntime) ConnectTeam(name string) {
	tc, err := c.Cfg.GetTeamConfig(name)
	if err != nil {
		c.PaneController.GetStatusPane().Logf("ERROR", "unable to locate team config for %s: %v", name, err)
		return
	}
	conn, err := connection.New(tc, c.rtmEvent)
	if err != nil {
		c.PaneController.GetStatusPane().Logf("ERROR", "unable to connect to %s: %v -- %+v", name, err, tc)
		return
	}
	c.PaneController.GetStatusPane().Log(tc.Name, "Connected..")
	c.Lock()
	c.teams = append(c.teams, conn)
	c.currentTeam = len(c.teams) - 1
	c.Unlock()
	c.PaneController.UpdateStatusBar()
}

func (c *AppRuntime) GetActiveTeam() *connection.Connection {
	return c.teams[c.currentTeam]
}

func (c *AppRuntime) TeamCount() int {
	return len(c.teams)
}

func (c *AppRuntime) GetTeam(teamId string) *connection.Connection {
	for _, e := range c.teams {
		if e.User.TeamID == teamId {
			return e
		}
	}
	return nil
}

func (c *AppRuntime) ActivateTeam(idx int) {
	if c.currentTeam == idx {
		c.PaneController.GetStatusPane().Logf(constants.Name, "Active team is already %s", c.GetActiveTeam().Name)
	} else {
		c.Lock()
		c.currentTeam = idx
		c.PaneController.GetStatusPane().Logf(constants.Name, "Active team is now %s", c.GetActiveTeam().Name)
		c.Unlock()
		c.PaneController.UpdateStatusBar()
	}
}

func (c *AppRuntime) ActivateNextTeam() {
	n := (c.currentTeam + 1) % len(c.teams)
	c.ActivateTeam(n)
}

func (c *AppRuntime) AddTicker(th TickHandler) {
	c.tickHandlers = append(c.tickHandlers, th)
}

func (c *AppRuntime) Run() error {
	c.ticker = time.NewTicker(1 * c.tickSize)
	go c.tickDispatcher(c.ticker.C)
	go c.postStartupConfiguration()
	return c.App.Run()
}

func (c *AppRuntime) Debugf(format string, args ...interface{}) {
	c.PaneController.GetStatusPane().Logf("DEBUG", format, args...)
}
func (c *AppRuntime) Debug(message string) {
	c.PaneController.GetStatusPane().Log("DEBUG", message)
}

func (c *AppRuntime) Stop() {
	c.ticker.Stop()
	c.App.Stop()
}

func (c *AppRuntime) postStartupConfiguration() {
	if c.Cfg.Debug {
		c.Debug("Post start-up Config")
	}
	for _, team := range c.Cfg.Teams {
		if team.AutoConnect {
			if c.Cfg.Debug {
				c.Debugf("Should Auto join Team: %s", team.Name)
			}
			cmd := &connectCommand{source: c.PaneController.GetStatusPane(), args: []string{team.Name}}
			cmd.Execute(c)
		}
	}
}

// function that sends ticks to all registered tickHandlers.
func (c *AppRuntime) tickDispatcher(ticks <-chan time.Time) {
	for {
		select {
		case tick := <-ticks:
			for _, f := range c.tickHandlers {
				go f.OnTick(tick, c.tickSize, c)
			}
		}
	}
}

// function passed down to PaneController to allow for queueing
// of updates from individual panes.
func (c *AppRuntime) queueUpdateDraw(f func()) {
	c.App.QueueUpdateDraw(f)
}

// function passed on to PaneController to allow for the handling
// of input and Commands in "Application Context"
func (c *AppRuntime) onInput(p *pane.Pane, input string) {
	if c.Cfg.Debug {
		c.Debugf("Got input: '%s'", input)
	}
	if input[0] == '/' {
		cmd, err := NewCommand(input, p)
		if err != nil {
			p.Log("ERROR", err.Error())
		}
		if cmd != nil {
			cmd.Execute(c)
		}
	} else {
		team := c.GetTeam(p.TeamId)
		if team != nil { // should only be nil on console-pane
			err := team.SendMessage(p.Channel.ID, input)
			if err != nil {
				p.Logf("ERROR", "unable to send message at this time: %v", err)
			}
		}
	}
}

// function passed on to  PaneController to allow for the handling of
// "special key" input from input fields.
func (c *AppRuntime) inputCapture(p *pane.Pane, event *tcell.EventKey) *tcell.EventKey {
	//c.PaneController.GetStatusPane().Logf("Key: %+v", event)
	switch event.Key() {
	case 24: // ^X
		c.ActivateNextTeam()
		return nil

	case 260: // left
		if event.Modifiers() == tcell.ModAlt {
			c.PaneController.SetPrevActive()
			return nil
		} else {
			return event
		}
	case 259: // right
		if event.Modifiers() == tcell.ModAlt {
			c.PaneController.SetNextActive()
			return nil
		} else {
			return event
		}
	default:
		return event
	}
}
