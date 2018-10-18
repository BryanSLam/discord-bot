package commands

import (
	"fmt"
	"strings"

	"github.com/BryanSLam/discord-bot/config"
	"github.com/BryanSLam/discord-bot/util"
	dg "github.com/bwmarrin/discordgo"
	iex "github.com/jonwho/go-iex"
)

func stock(s *dg.Session, m *dg.MessageCreate) {
	slice := strings.Split(m.Content, " ")
	ticker := slice[1]
	iexClient := iex.NewClient()
	logger := util.Logger{Session: s, ChannelID: config.GetConfig().BotLogChannelID}

	logger.Info("Fetching stock info for " + ticker)
	quote, err := iexClient.Quote(ticker, true)

	if err != nil {
		rds, iexErr := iexClient.RefDataSymbols()
		if iexErr != nil {
			logger.Trace("IEX request failed. Message: " + iexErr.Error())
			s.ChannelMessageSend(m.ChannelID, iexErr.Error())
			return
		}

		fuzzySymbols := util.FuzzySearch(ticker, rds.Symbols)

		if len(fuzzySymbols) > 0 {
			fuzzySymbols = fuzzySymbols[:len(fuzzySymbols)%10]
			outputJSON := util.FormatFuzzySymbols(fuzzySymbols)
			s.ChannelMessageSend(m.ChannelID, outputJSON)
			return
		}
	}

	if quote == nil {
		logger.Trace(fmt.Sprintf("nil quote from ticker: %s", ticker))
	}

	message := util.FormatQuote(quote)
	s.ChannelMessageSend(m.ChannelID, message)
}
