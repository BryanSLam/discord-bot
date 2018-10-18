package commands

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/BryanSLam/discord-bot/config"
	"github.com/BryanSLam/discord-bot/util"
	dg "github.com/bwmarrin/discordgo"
)

func remindme(s *dg.Session, m *dg.MessageCreate) {
	logger := util.Logger{Session: s, ChannelID: config.GetConfig().BotLogChannelID}

	logger.Info("Received reminder request `" + m.Content + "`")

	submatches := remindmeRegex.FindStringSubmatch(m.Content)
	logger.Info(fmt.Sprintf("Parsed reminder request %#v\n", submatches))

	if len(submatches) != 4 {
		s.ChannelMessageSend(m.ChannelID, "expected command format: `!remindme 2 days \"message!\"`")
		logger.Trace("Received incorrect arguments for reminder on `" + m.Content + "`")
		return
	}

	timeValue, _ := strconv.Atoi(submatches[1])
	timeUnit := submatches[2]
	reminderMessage := submatches[3]

	spec, err := cronSpec(timeValue, timeUnit)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	reminderCron.AddFunc(spec, func() {
		s.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("%s REMINDER! %s\n", m.Author.Mention(), reminderMessage))
	})

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s your reminder is set\n", m.Author.Mention()))
}

func cronSpec(value int, unit string) (string, error) {
	var spec time.Time

	now := time.Now().Local()
	switch unit {
	case "year":
		spec = now.Add(time.Hour * 8760 * time.Duration(value))
	case "years":
		spec = now.Add(time.Hour * 8760 * time.Duration(value))
	case "month":
		spec = now.Add(time.Hour * 730 * time.Duration(value))
	case "months":
		spec = now.Add(time.Hour * 730 * time.Duration(value))
	case "day":
		spec = now.Add(time.Hour * 24 * time.Duration(value))
	case "days":
		spec = now.Add(time.Hour * 24 * time.Duration(value))
	case "hour":
		spec = now.Add(time.Hour * time.Duration(value))
	case "hours":
		spec = now.Add(time.Hour * time.Duration(value))
	case "minute":
		spec = now.Add(time.Minute * time.Duration(value))
	case "minutes":
		spec = now.Add(time.Minute * time.Duration(value))
	default:
		return "", errors.New("Unable to parse time for reminder")
	}

	seconds := spec.Second()
	minutes := spec.Minute()
	hours := spec.Hour()
	day := spec.Day()
	month := spec.Month()
	weekday := spec.Weekday()

	return fmt.Sprintf("%d %d %d %d %d %d", seconds, minutes, hours, day, month, weekday), nil
}
