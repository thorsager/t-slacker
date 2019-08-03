package runtime

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/thorsager/t-slacker/config"
	"github.com/thorsager/t-slacker/connection"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	userPattern   = regexp.MustCompile(`(<@[A-Z0-9]+>)`)
	userIdPattern = regexp.MustCompile(`^<@([A-Z0-9]+)>$`)
)

func (c *AppRuntime) renderMessageEvent(msg *slack.MessageEvent) []byte {
	if msg.SubType == "message_replied" {
		return []byte("\nTODO: handle replies")
	} else {
		team := c.GetTeam(msg.Team)
		t := asTime(msg.Timestamp)
		uname := msg.User
		color := "-"
		if team.Config.Colorize {
			color = colorize(uname)
		}
		if u, err := team.UserLookup(uname); err == nil {
			uname = u.Name
		}
		parsedText := parseUsers(team, msg.Text)
		parsedText += renderFiles(msg.Files)
		return []byte(fmt.Sprintf("\n%.2d:%.2d [#666666]<[-][%s]%s[-][#666666]>[-] %s", t.Hour(), t.Minute(),
			color, uname, parsedText))
	}
}

func (c *AppRuntime) renderMessage(teamId string, msg slack.Message) []byte {
	team := c.GetTeam(teamId)
	t := asTime(msg.Timestamp)
	uname := msg.User
	color := "-"
	if team.Config.Colorize {
		color = colorize(uname)
	}
	if u, err := team.UserLookup(uname); err == nil {
		uname = u.Name
	}
	parsedText := parseUsers(team, msg.Text)
	parsedText += renderFiles(msg.Files)
	return []byte(fmt.Sprintf("\n%.2d:%.2d [#666666]<[-][%s]%s[-][#666666]>[-] %s", t.Hour(), t.Minute(), color, uname, parsedText))
}

func renderFiles(files []slack.File) (rendered string) {
	rendered = ""
	if files != nil && len(files) > 0 {
		rendered += "\n[#888888]  Attachments:[-]"
		for _, file := range files {
			if file.Mimetype == "application/vnd.slack-docs" {
				rendered += fmt.Sprintf("\n    - [#888888]%s[-]", file.Permalink)
			} else {
				rendered += fmt.Sprintf("\n    - [#888888]%s[-]", file.URLPrivate)
			}
		}
	}
	return rendered
}

func parseUsers(team *connection.Connection, msg string) string {
	for _, usr := range userPattern.FindAllString(msg, -1) {

		uid := isolateId(usr)
		color := "-"
		if team.Config.ColorizeInline {
			color = colorize(uid)
		}
		name := lookupUser(team, uid)
		msg = strings.Replace(msg, usr, fmt.Sprintf("[#666666]@[-][%s]%s[-]", color, name), -1)
	}
	return msg
}
func isolateId(userSlug string) string {
	uid := userIdPattern.FindStringSubmatch(userSlug)
	return uid[1]
}

func lookupUser(team *connection.Connection, uid string) string {
	if team == nil {
		return uid
	}
	user, err := team.UserLookup(uid)
	if err != nil {
		return uid
	}
	return user.Name
}

func colorize(s string) string {
	var sum int
	for _, i := range []byte(s) {
		sum += int(i)
	}
	return config.ColorizeColors[sum%len(config.ColorizeColors)]
}

func asTime(timeStamp string) time.Time {
	floatTime, err := strconv.ParseFloat(timeStamp, 64)
	if err != nil {
		floatTime = 0.0
	}
	intTime := int64(floatTime)
	return time.Unix(intTime, 0)
}
