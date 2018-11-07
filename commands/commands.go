package commands

import (
	"regexp"
)

var (
	commandRegex      = regexp.MustCompile(`(?i)^![a-z]+`)
	pingRegex         = regexp.MustCompile(`(?i)^!ping$`)
	stockRegex        = regexp.MustCompile(`(?i)^!stock [a-z\.]+$`)
	erRegex           = regexp.MustCompile(`(?i)^!er [a-z]+$`)
	wizdaddyRegex     = regexp.MustCompile(`(?i)^!wizdaddy$`)
	coinRegex         = regexp.MustCompile(`(?i)^!coin [a-z]+$`)
	remindmeRegex     = regexp.MustCompile(`(?i)^!remindme [ a-z".]*[ 0-9/]*`)
	remindmeDateRegex = regexp.MustCompile(`(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])/(\d\d)`)
)
