package runtime

import (
	"github.com/thorsager/t-slacker/connection"
	"github.com/thorsager/t-slacker/constants"
	"github.com/thorsager/t-slacker/pane"
	"strings"
)

func (c *AppRuntime) tabCapture(p *pane.Pane, input string) string {
	segments := strings.Split(input, " ")
	ltok := segments[len(segments)-1]
	if ltok == "" {
		return input
	}
	niToken := ltok[1:]
	if c.tab.Indicator != ltok[0] || niToken != c.tab.LastReturned {
		c.tab.Indicator = ltok[0]
		c.tab.Count = 0
		c.tab.Token = niToken
	}
	if niToken == c.tab.LastReturned {
		c.tab.Count = c.tab.Count + 1
	}

	var team *connection.Connection
	var output string
	if p.TeamId == constants.ConsoleTeamChannelId {
		team = c.GetActiveTeam()
	} else {
		team = c.GetTeam(p.TeamId)
	}
	switch c.tab.Indicator {
	case constants.UserIndicatorChar:
		userNames, err := team.FindUserNamesStartingWith(c.tab.Token)
		if err != nil || len(userNames) < 1 {
			output = input
			break
		}
		idx := c.tab.Count % len(userNames)
		newToken := userNames[idx]
		output = join(segments, string(c.tab.Indicator)+newToken)
		c.tab.LastReturned = newToken //store for repeat detection
	case constants.ChannelIndicatorChar:
		channelNames, err := team.FindChannelNamesStartingWith(c.tab.Token)
		if err != nil || len(channelNames) < 1 {
			output = input
			break
		}
		idx := c.tab.Count % len(channelNames)
		newToken := channelNames[idx]
		output = join(segments, string(c.tab.Indicator)+newToken)
		c.tab.LastReturned = newToken //store for repeat detection
	default:
		output = input
	}
	return output
}

func join(segments []string, newToken string) string {
	var res string
	if len(segments) > 1 {
		res = strings.Join(segments[:len(segments)-1], " ")
		res += " "
	}
	return res + newToken
}
