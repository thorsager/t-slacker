package runtime

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
	"strings"
)

type listCommand struct {
	source *pane.Pane
	args   []string
}

func (c *listCommand) Execute(ctx *AppRuntime) {
	types := []string{connection.TPrivateChannel, connection.TPublicChannel}
	if len(c.args) > 0 && strings.ToUpper(c.args[0]) == "-ALL" {
		types = append(types, connection.TIm, connection.TMpim)
	}
	if c.source == ctx.PaneController.GetStatusPane() {
		team := ctx.GetActiveTeam()
		list, err := team.GetConversationsCaching(types...)
		if err != nil {
			ctx.PaneController.GetActive().Logf("ERROR", "unable to fetch channel-list: %v", err)
		}
		ctx.PaneController.GetStatusPane().Log(team.Name, "Channel list")
		for _, e := range list {
			ctx.PaneController.GetStatusPane().Log(team.Name, format(e, team))
		}
		ctx.PaneController.GetStatusPane().Log(team.Name, "End of Channel list.")
	} else {
		ctx.PaneController.GetActive().Log(constants.Name, "list command only support in status-pane")
	}
}

func format(c slack.Channel, team *connection.Connection) string {
	if c.IsIM {
		return formatIM(c, team)
	} else if c.IsGroup {
		return formatGroup(c, team)
	} else {
		return formatChannel(c)
	}
}

func formatChannel(c slack.Channel) string {
	return fmt.Sprintf("%s%s %d [%s %d:%d] %s",
		constants.ChannelIndicator,
		c.Name, c.NumMembers,
		channelMod(c), c.UnreadCount, c.UnreadCountDisplay,
		c.Topic.Value)
}

func formatGroup(i slack.Channel, team *connection.Connection) string {
	return fmt.Sprintf("%s%s %d [%s %d:%d] %s",
		constants.GroupIndicator,
		i.Name, i.NumMembers,
		channelMod(i), i.UnreadCount, i.UnreadCountDisplay,
		i.Topic.Value)
}

func formatIM(i slack.Channel, team *connection.Connection) string {
	name := i.User
	user, err := team.UserLookup(i.User)
	if err == nil {
		name = user.Name
	}
	return fmt.Sprintf("@%s %d [%s %d:%d] %s (%+v)",
		name, i.NumMembers,
		channelMod(i), i.UnreadCount, i.UnreadCountDisplay,
		i.Topic.Value, i)
}

func channelMod(c slack.Channel) string {
	mod := ""
	if c.IsArchived {
		mod += "a"
	}
	if c.IsChannel {
		mod += "c"
	}
	if c.IsGeneral {
		mod += "G"
	}
	if c.IsGroup {
		mod += "g"
	}
	if c.IsIM {
		mod += "i"
	}
	if c.IsPrivate {
		mod += "p"
	}
	if c.IsExtShared {
		mod += "e"
	}
	if c.IsOrgShared {
		mod += "o"
	}
	if c.IsOpen {
		mod += "O"
	}
	if c.IsMember {
		mod += "M"
	}
	if mod != "" {
		return "+" + mod
	} else {
		return ""
	}
}
