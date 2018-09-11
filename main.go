package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/BryanSLam/discord-bot/commands"
	"github.com/BryanSLam/discord-bot/datasource"
	"github.com/BryanSLam/discord-bot/util"
	iex "github.com/jonwho/go-iex"
	"github.com/robfig/cron"

	"github.com/bwmarrin/discordgo"
	"github.com/tkanos/gonfig"
)

type botConfig struct {
	CoinAPIURL            string
	InvalidCommandMessage string
	WizdaddyURL           string
}

// Variables to initialize
var (
	token          string
	config         botConfig
	iexClient      *iex.Client
	reminderClient commands.Reminder
)

func init() {
	// Run the program with `go run main.go -t <token>`
	// flag.Parse() will assign to token var
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()

	// If no value was provided from flag look for env var BOT_TOKEN
	if token == "" {
		token = os.Getenv("BOT_TOKEN")
	}

	// Initialize iexClient with new client
	iexClient = iex.NewClient()

	// Initalize new reminder goroutine
	reminderClient = commands.NewReminder("")

	// Use gonfig to fetch the config variables from config.json
	err := gonfig.GetConf("config.json", &config)
	if err != nil {
		fmt.Println("error fetching config values", err)
		return
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// 5 AM everyday Monday - Friday
	go func() {
		c := cron.New()
		c.AddFunc("0 5 * * 1-5", func() {
			fmt.Println("test")
			err := reminderRoutine(dg)
			if err != nil {
				fmt.Println(err)
			}
		})
		c.Start()

	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "!ping" reply with "pong!"
	if match, _ := regexp.MatchString("!ping", m.Content); match {
		s.ChannelMessageSend(m.ChannelID, "pong!")
	}

	if match, _ := regexp.MatchString("![a-zA-Z]+[ a-zA-Z\"]*[ 0-9/]*", m.Content); match {
		slice := strings.Split(m.Content, " ")

		if action, _ := regexp.MatchString("(?i)^!stock$", slice[0]); action {
			ticker := slice[1]
			quote, err := iexClient.Quote(ticker, true)

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())

				rds, iexErr := iexClient.RefDataSymbols()
				if iexErr != nil {
					s.ChannelMessageSend(m.ChannelID, iexErr.Error())
				}

				fuzzySymbols := util.FuzzySearch(ticker, rds.Symbols)

				if len(fuzzySymbols) > 0 {
					fuzzySymbols = fuzzySymbols[:len(fuzzySymbols)%10]
					outputJSON := util.FormatFuzzySymbols(fuzzySymbols)
					s.ChannelMessageSend(m.ChannelID, outputJSON)
				}
				return
			}

			outputJSON := util.FormatQuote(quote)

			s.ChannelMessageSend(m.ChannelID, outputJSON)
		} else if action, _ := regexp.MatchString("(?i)^!er$", slice[0]); action {
			ticker := slice[1]
			earnings, err := iexClient.Earnings(ticker)

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			outputJSON := util.FormatEarnings(earnings)

			s.ChannelMessageSend(m.ChannelID, outputJSON)
		} else if action, _ := regexp.MatchString("(?i)^!remindme$", slice[0]); action {
			messageArr := strings.Split(m.Content, "\"")
			if len(messageArr) != 3 {
				s.ChannelMessageSend(m.ChannelID, "Reminder messages must be surrounded by quotes \"{message}\" ")
				return
			}
			// We store the person who sent the message as well as the channel id into the redis cache so we know where and who to contact later
			message := m.ChannelID + "~*" + m.Author.Mention() + ": " + messageArr[1]
			date := slice[len(slice)-1]
			match, _ := regexp.MatchString("(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])/(\\d\\d)", date)
			if match == false {
				s.ChannelMessageSend(m.ChannelID, "Invalid date given loser")
				return
			}

			// Commenting out the date check for now, weird behavior where you get blocked for
			// Setting a reminder for the next day

			// dateCheck, _ := time.Parse("01/02/06", date)
			// if time.Until(dateCheck) < 0 {
			// 	s.ChannelMessageSend(m.ChannelID, "Date has already passed ya fuck")
			// 	return
			// }

			err := reminderClient.Add(message, date)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			s.ChannelMessageSend(m.ChannelID, "Reminder Set!")
		} else if action, _ := regexp.MatchString("(?i)^!wizdaddy", slice[0]); action {
			resp, err := http.Get(config.WizdaddyURL)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Daddy is down")
				return
			}

			var daddyResponse datasource.WizdaddyResponse
			if err = json.NewDecoder(resp.Body).Decode(&daddyResponse); err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID,
				fmt.Sprintf("Daddy says to buy: %s %s %s %s", daddyResponse.Symbol,
					daddyResponse.StrikePrice, daddyResponse.ExpirationDate, daddyResponse.Type))
		} else if action, _ := regexp.MatchString("(?i)^!coin$", slice[0]); action {
			ticker := strings.ToUpper(slice[1])
			coinURL := config.CoinAPIURL + ticker + "&tsyms=USD"

			resp, err := http.Get(coinURL)

			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			coin := datasource.Coin{Symbol: ticker}

			if err = json.NewDecoder(resp.Body).Decode(&coin); err != nil || coin.Response == "Error" {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			s.ChannelMessageSend(m.ChannelID, coin.OutputJSON())
			defer resp.Body.Close()
		} else {
			s.ChannelMessageSend(m.ChannelID, config.InvalidCommandMessage)
		}
	}
}

// Function run during the daily reminder check
func reminderRoutine(s *discordgo.Session) error {
	output, err := reminderClient.Get(time.Now().Format("01/02/06"))
	if err != nil {
		return err
	}
	for _, reminder := range output {
		cacheEntry := strings.Split(reminder, "~*")
		channel := cacheEntry[0]
		s.ChannelMessageSend(channel, cacheEntry[1])
	}
	return nil
}
