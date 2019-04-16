package commands

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/BryanSLam/discord-bot/config"
	"github.com/BryanSLam/discord-bot/util"
	dg "github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
	"github.com/robfig/cron"
)

const (
	dateFormat        string = "1/_2/06"
	redisDateFormat   string = "01/02/06"
	watchlistRedisKey string = "watchlist"
)

type work func(s *dg.Session, m *dg.MessageCreate)

var (
	token        string
	redisClient  *redis.Client
	cronner      *cron.Cron
	pst, _       = time.LoadLocation("America/Los_Angeles")
	commandRegex = regexp.MustCompile(`(?i)^![\w]+[\w ".]*[ 0-9/]*$`)
	commands     []command
)

var (
	coinAPIURL            = config.GetConfig().CoinAPIURL
	wizdaddyURL           = config.GetConfig().WizdaddyURL
	invalidCommandMessage = config.GetConfig().InvalidCommandMessage
	botLogChannelID       = config.GetConfig().BotLogChannelID
	earningsWhisperURL    = config.GetConfig().EarningsWhisperURL
)

var (
	ping           = newPingCommand()
	stock          = newStockCommand()
	er             = newErCommand()
	wizdaddy       = newWizdaddyCommand()
	coin           = newCoinCommand()
	remindme       = newRemindmeCommand()
	watchlist      = newWatchlistCommand()
	clearwatchlist = newClearWatchlistCommand()
	news           = newNewsCommand()
	nexter         = newNextErCommand()
)

func init() {
	token = os.Getenv("BOT_TOKEN")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	cronner = cron.NewWithLocation(pst)
	cronner.Start()

	commands = append(commands, ping, stock, er, wizdaddy, coin, remindme, watchlist, clearwatchlist,
		news, nexter)
}

func Commander() func(s *dg.Session, m *dg.MessageCreate) {
	return func(s *dg.Session, m *dg.MessageCreate) {

		if commandRegex.MatchString(m.Content) {
			// Ignore all messages created by the bot itself
			// This isn't required in this specific example but it's a good practice.
			if m.Author.ID == s.State.User.ID {
				return
			}

			for _, cmd := range commands {
				if cmd.match(m.Content) {
					go safelyDo(cmd.fn, s, m)
					return
				}
			}

			s.ChannelMessageSend(m.ChannelID, invalidCommandMessage)
		}
	}
}

func safelyDo(fn work, s *dg.Session, m *dg.MessageCreate) {
	logger := util.Logger{Session: s, ChannelID: botLogChannelID}

	// defer'd funcs will execute before return even if runtime panic
	defer func() {
		// use recover to catch panic so bot doesn't shutdown
		if err := recover(); err != nil {
			logger.Send(util.MentionMaintainers() + " an error has occurred")
			logger.Warn(fmt.Sprintln("function", util.FuncName(fn), "failed:", err))
		}
	}()
	// perform work
	fn(s, m)
}
