package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/BryanSLam/discord-bot/commands"
	"github.com/robfig/cron"

	"github.com/bwmarrin/discordgo"
)

// Variables to initialize
var (
	token string
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

	if token == "" {
		log.Fatalln("discord token must be supplied")
	}
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("error creating Discord session,", err)
	}

	// Register handlers for discordgo
	dg.AddHandler(commands.Commander())

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection,", err)
	}

	// 5 AM everyday Monday - Friday
	go func() {
		c := cron.New()
		c.AddFunc("@midnight", func() {
			log.Println("cron running!")
			// TODO: main starting to look messy again
			//
			// move behaviors into separate pkgs
			err := commands.ReminderRoutine(dg)
			if err != nil {
				log.Println(err)
			}
		})
		c.Start()
	}()

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
