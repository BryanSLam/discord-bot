package commands

import (
	"fmt"
	"strings"
	"time"

	dg "github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

/*
 * Set up redis table in a way where the keys are the dates of when to pull a reminder
 * and the value is an array of events that need to be reminded on that date
 */
type Reminder struct {
	Client *redis.Client
}

var (
	reminderClient Reminder
)

func init() {
	// Initalize new reminder goroutine
	reminderClient = NewReminder("")
}

// TODO: decide on pkg API. not sure if this stuff should be public or not
//
// where else will this be used?
func NewReminder(url string) Reminder {
	var (
		storeurl string
		client   *redis.Client
	)
	// For mock testing
	if url != "" {
		storeurl = url
	} else {
		// Replace with the config from the charts PR
		storeurl = "redis:6379"
	}
	client = redis.NewClient(&redis.Options{
		Addr:     storeurl,
		Password: "", // No password set
		DB:       0,  // Use default DB
	})
	return Reminder{
		Client: client,
	}
}

// Add adds the new reminder as an entry into the redis table, we assume validation is
// done when the message was recieved and the date should be in the form **/**/**
// Append the reminder to the existing list if it exists, if not create a new list and add it
func (r *Reminder) Add(message string, date string) error {
	var (
		sb  strings.Builder
		val string
	)
	val, err := r.Client.Get(date).Result()
	if err != nil {
		if err.Error() != "redis: nil" {
			return err
		}
	}
	if val != "" {
		sb.WriteString(val)
		sb.WriteString("::")
	}
	sb.WriteString(message)

	untilDate, err := time.Parse("01/02/06", date)
	if err != nil {
		return err
	}

	// Might need to add a buffer for the duration that the redis entry exists so we have
	// time to get the value and output reminder
	timeUntil := time.Until(untilDate) + 1
	err = r.Client.Set(date, sb.String(), timeUntil).Err()
	if err != nil {
		fmt.Println("Add set error")
		return err
	}
	return nil
}

// Get will run on a daily cron job to fetch the raw messages from the redis cache and send to the
// parser to be formatted before being broadcasted. The expected key is the current date
func (r *Reminder) Get(date string) ([]string, error) {
	messages, err := r.Client.Get(date).Result()
	if err != nil {
		if err.Error() != "redis: nil" {
			return nil, err
		}
		return nil, nil
	}
	output := strings.Split(messages, "::")

	// Delete the entries in the redis cache after it's gotten in case the duration doesn't expire
	r.Client.Del(date)
	return output, nil
}

func remindme(s *dg.Session, m *dg.MessageCreate) {
	messageArr := strings.Split(m.Content, "\"")
	if len(messageArr) != 3 {
		s.ChannelMessageSend(m.ChannelID, "Reminder messages must be surrounded by quotes \"{message}\" ")
		return
	}

	// We store the person who sent the message as well as the channel id into the redis cache
	// so we know where and who to contact later
	message := m.ChannelID + "~*" + m.Author.Mention() + ": " + messageArr[1]
	slice := strings.Split(m.Content, " ")
	date := slice[len(slice)-1]
	if !remindmeDateRegex.MatchString(date) {
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
}

// Function run during the daily reminder check
func ReminderRoutine(s *dg.Session) error {
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
