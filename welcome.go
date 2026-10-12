package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	api "github.com/quackduck/devzat/devzatapi"
)

var joinRe = regexp.MustCompile(`--> (.+?) has joined the chat`)

const defaultGuide = `Hey {{name}}, welcome to the team chat! 👋

This is an SSH chat server. A few things to get you going:

• ` + "`help`" + ` lists commands; ` + "`cd`" + ` lists rooms and ` + "`cd #room`" + ` joins one.
• ` + "`gh pr <owner/repo> <n>`" + ` or ` + "`gh issue …`" + ` pulls up a PR or issue right here.
• ` + "`gif <query>`" + ` posts a GIF.
• ` + "`remind <when> <message>`" + ` (e.g. ` + "`remind 2h review #292`" + `) pings you later.
• ` + "`@name`" + ` mentions someone (Tab completes names); ` + "`=name <msg>`" + ` sends a DM.
• ` + "`bell on|off|all`" + ` controls notification bells; ` + "`nick`" + `, ` + "`pronouns`" + `, ` + "`tz`" + `, ` + "`theme`" + `, ` + "`color`" + ` personalize your session.

Ping an admin if you need anything. Glad to have you aboard!`

func loadGuide() string {
	path := os.Getenv("WELCOME_FILE")
	if path == "" {
		return defaultGuide
	}
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not read WELCOME_FILE, using default guide:", err)
		return defaultGuide
	}
	return string(b)
}

func parseJoinName(text string) string {
	m := joinRe.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(m[1])
}

func startWelcome(s *api.Session) error {
	guide := loadGuide()
	msgCh, _, err := s.RegisterListener(false, false, "has joined the chat", api.WithSystemMessages())
	if err != nil {
		return err
	}

	// `welcome` re-sends the guide on demand (useful for testing and for re-reading).
	err = s.RegisterCmd("welcome", "[@user]", "Show the welcome guide again", func(cmd api.CmdCall, e error) {
		if e != nil {
			return
		}
		target := cmd.From
		if f := strings.Fields(cmd.Args); len(f) > 0 {
			target = strings.TrimPrefix(f[0], "@")
		}
		text := strings.ReplaceAll(guide, "{{name}}", target)
		if sendErr := s.SendMessage(api.Message{Room: cmd.Room, From: welcomeName, DMTo: target, Data: text}); sendErr != nil {
			_ = s.SendMessage(api.Message{
				Room: cmd.Room, DMTo: cmd.From,
				Data: "Couldn't DM " + target + " — are they online in this room?",
			})
		}
	})
	if err != nil {
		return err
	}

	go func() {
		var mu sync.Mutex
		welcomed := map[string]bool{}
		for m := range msgCh {
			name := parseJoinName(m.Data)
			if name == "" {
				continue
			}
			mu.Lock()
			already := welcomed[name]
			welcomed[name] = true
			mu.Unlock()
			if already {
				continue
			}
			text := strings.ReplaceAll(guide, "{{name}}", name)
			if err := s.SendMessage(api.Message{Room: m.Room, From: welcomeName, DMTo: name, Data: text}); err != nil {
				fmt.Fprintln(os.Stderr, "welcome DM failed:", err)
			}
		}
	}()
	return nil
}
