package runtime

import (
	"fmt"
	"github.com/thorsager/t-slacker/pane"
	"strings"
)

type Command interface {
	Execute(ctx *AppRuntime)
}

func NewCommand(s string, source *pane.Pane) (Command, error) {
	if len(s) < 1 || s[0] != '/' {
		return nil, fmt.Errorf("invalid command string: '%s'", s)
	}
	tokens := tokenize(s[1:])

	switch strings.ToUpper(tokens[0]) {
	case "QUIT", "Q":
		return &quitCommand{}, nil
	case "HELP", "H":
		return &helpCommand{}, nil
	case "WINDOW", "WD", "WND", "W":
		return &windowCommand{args: tokens[1:], source: source}, nil
	case "CONNECT", "CN":
		return &connectCommand{args: tokens[1:], source: source}, nil
	case "LIST", "LS":
		return &listCommand{args: tokens[1:], source: source}, nil
	case "JOIN", "J":
		return &joinCommand{args: tokens[1:], source: source}, nil
	case "PART", "LEAVE":
		return &partCommand{args: tokens[1:], source: source}, nil
	case "PRIVMSG", "MSG":
		return &privMsgCommand{args: tokens[1:], source: source}, nil
	default:
		return nil, fmt.Errorf("command not found: '%s'", strings.ToUpper(tokens[0]))
	}
}

func tokenize(src string) []string {
	//var s scanner.Scanner
	//s.Init(strings.NewReader(src))
	//slice := make([]string, 0, 5)
	//for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
	//	slice = append(slice, s.TokenText())
	//}
	//return slice
	return strings.Split(src, " ")
}
