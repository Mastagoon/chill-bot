package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Record struct {
	Count             int
	OldestMessageTime time.Time
	WordCount         int
	IsMuted           bool
}

type List struct {
	records []Record
}

var list map[string]*Record
var insults = []string{"يا اكحل العينين", "يا.. يا.. يا كثير الكلام!", "ايها الأرعن", "يا مغفل", "يا ثرثار", "يا مزعج", "يا متطفل", "يا ضعيف الإرادة", "يا أحمق", "يا متعجرف"}
var MAX_MESSAGES = 10

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	token := os.Getenv("BOT_TOKEN")

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
	}
	discord.AddHandler(messageCreate)
	discord.AddHandler(messageEdit)
	discord.AddHandler(ready)
	discord.Identify.Intents = discordgo.IntentsGuildMessages
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	list = make(map[string]*Record)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.ID == "679348712472051715" {
		if st := strings.Split(m.Content, " "); st[0] == "!max" {
			i, err := strconv.Atoi(st[1])
			if err == nil {
				MAX_MESSAGES = i
				s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("تم تغيير الحد الأقصى للرسائل إلى %d", i), m.Reference())
			}
		}
	}
	length := len(strings.Split(m.Content, " "))
	if len(m.Attachments) > 0 {
		length += 20
	}
	length += len(m.Content) / 5
	var r *Record
	var ok bool
	if r, ok = list[m.Author.ID]; !ok {
		list[m.Author.ID] = &Record{Count: 1, OldestMessageTime: m.Timestamp, WordCount: length}
		r = list[m.Author.ID]
	} else {
		if r.IsMuted {
			if r.OldestMessageTime.Before(time.Now().Add(-10 * time.Minute)) {
				r.IsMuted = false
			} else {
				r.OldestMessageTime = m.Timestamp
			}
		}
		if r.OldestMessageTime.Before(time.Now().Add(-10 * time.Minute)) { // reset
			r.Count = 1
			r.WordCount = length
		} else { //not muted, not reset
			r.Count++
			r.WordCount += length
		}
	}
	if r.Count > MAX_MESSAGES || r.WordCount > 20 || r.IsMuted {
		r.IsMuted = true
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func messageEdit(s *discordgo.Session, m *discordgo.MessageUpdate) {
	messageCreate(s, &discordgo.MessageCreate{Message: m.Message})
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Println(fmt.Sprintf("Connected as %s#%s", event.User.Username, event.User.Discriminator))
}
