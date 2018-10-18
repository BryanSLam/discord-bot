package commands

import (
	"regexp"

	"github.com/robfig/cron"
)

var (
	commandRegex  = regexp.MustCompile(`(?i)^![a-z]+`)
	pingRegex     = regexp.MustCompile(`(?i)^!ping$`)
	stockRegex    = regexp.MustCompile(`(?i)^!stock [a-z\.]+$`)
	erRegex       = regexp.MustCompile(`(?i)^!er [a-z]+$`)
	wizdaddyRegex = regexp.MustCompile(`(?i)^!wizdaddy$`)
	coinRegex     = regexp.MustCompile(`(?i)^!coin [a-z]+$`)
	remindmeRegex = regexp.MustCompile(`(?i)^!remindme (?P<TimeValue>[1-9]+[0-9]*) (?P<TimeUnit>minutes?|hours?|days?|months?|years?) (?P<Message>".+")$`)

	reminderCron *cron.Cron
)

func init() {
	reminderCron = cron.New()
	reminderCron.Start()
}
