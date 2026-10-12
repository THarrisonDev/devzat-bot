package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "github.com/quackduck/devzat/devzatapi"
)

func startRemind(s *api.Session) error {
	return s.RegisterCmd(
		"remind",
		"<when> <message>",
		"Be reminded of something later (e.g. `remind 2h check the deploy`)",
		func(cmd api.CmdCall, e error) {
			if e != nil {
				return
			}
			handleRemind(s, cmd)
		},
	)
}

func handleRemind(s *api.Session, cmd api.CmdCall) {
	parts := strings.SplitN(strings.TrimSpace(cmd.Args), " ", 2)
	if len(parts) < 2 {
		_ = s.SendMessage(api.Message{
			Room: cmd.Room, DMTo: cmd.From,
			Data: "Usage: `remind <when> <message>` — e.g. `remind 2h review the deploy` (units: s, m, h, d).",
		})
		return
	}

	dur, err := parseDuration(parts[0])
	if err != nil || dur <= 0 {
		_ = s.SendMessage(api.Message{
			Room: cmd.Room, DMTo: cmd.From,
			Data: fmt.Sprintf("Couldn't parse %q as a duration. Try `30m`, `2h`, `1d`, or `1d12h`.", parts[0]),
		})
		return
	}

	message := strings.TrimSpace(parts[1])
	room, from := cmd.Room, cmd.From

	time.AfterFunc(dur, func() {
		text := fmt.Sprintf("@%s reminder: %s", from, message)
		if err := s.SendMessage(api.Message{Room: room, From: remindName, Data: text}); err != nil {
			// The room may have been cleaned up; fall back to the main room.
			if err := s.SendMessage(api.Message{Room: "#main", From: remindName, Data: text}); err != nil {
				fmt.Fprintln(os.Stderr, "reminder delivery failed:", err)
			}
		}
	})

	_ = s.SendMessage(api.Message{
		Room: cmd.Room, DMTo: from,
		Data: fmt.Sprintf("Okay — I'll remind you in %s: %s", dur, message),
	})
}

var daysRe = regexp.MustCompile(`^(\d+)d(.*)$`)

// parseDuration is like time.ParseDuration but also supports days, e.g. "2d" or "1d12h".
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if m := daysRe.FindStringSubmatch(s); m != nil {
		days, err := strconv.Atoi(m[1])
		if err != nil {
			return 0, err
		}
		var rest time.Duration
		if m[2] != "" {
			rest, err = time.ParseDuration(m[2])
			if err != nil {
				return 0, err
			}
		}
		return time.Duration(days)*24*time.Hour + rest, nil
	}
	return time.ParseDuration(s)
}
