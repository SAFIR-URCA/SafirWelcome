package main

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"math/rand"
	"net/smtp"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	DiscordConfig DiscordConfig `json:"discord"`
	MailConfig    MailConfig    `json:"mail"`
}

type DiscordConfig struct {
	Token string `json:"token"`
}

type MailConfig struct {

	// Sender config
	From     string `json:"from"`
	Password string `json:"password"`

	// Server config
	SMTPHost string `json:"SMTPHost"`
	SMTPPort string `json:"SMTPPort"`
}

type Student struct {
	discordId string
	code      string
}

var students []Student
var config Config

func displayMenu(s *discordgo.Session, m *discordgo.MessageCreate) {
	var embed discordgo.MessageEmbed

	embed = getEmbed()
	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "S'enregistrer en tant qu'étudiant",
			Value:  "!server register prenom.nom@etudiant.univ-reims.fr",
			Inline: false,
		},

		{
			Name:   "Valider l'enregistrement",
			Value:  "!server verify leCodeReçuparMail",
			Inline: false,
		},
	}

	embed.Fields = fields

	_, err := s.ChannelMessageSendEmbed(m.ChannelID, &embed)
	if err != nil {
		logrus.Error(err)
	}
}

func getEmbed() discordgo.MessageEmbed {
	var embed discordgo.MessageEmbed
	var rootMeEmbedFooter discordgo.MessageEmbedFooter
	var rootMeEmbedThumbnail discordgo.MessageEmbedThumbnail

	embed.Color = 0x9fef00

	rootMeEmbedThumbnail.URL = "https://i.imgur.com/h7TH7bt.png"
	embed.Thumbnail = &rootMeEmbedThumbnail

	embed.Title = "SAFIR-URCA - Welcome"
	embed.Description = "SAFIR-URCA powered <:safir:674335843304865819>"

	rootMeEmbedFooter.Text = "Made with ❣️ by RavenXploit"
	embed.Footer = &rootMeEmbedFooter

	return embed
}

func handleDiscordMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!server" || m.Content == "!server help" {
		displayMenu(s, m)
	}

	if strings.Contains(m.Content, "!server register") {

		re := regexp.MustCompile("!server register ([\\w\\d.!#$%&'*+-/=?^_`{|}~]+@etudiant.univ-reims.fr)")
		match := re.FindStringSubmatch(m.Content)

		if match == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Cet email est invalide <@"+m.Author.ID+"> !")
			if err != nil {
				logrus.Error(err)
			}
		} else {
			rand.Seed(time.Now().UnixNano())
			code := randSeq(10)
			st := Student{
				discordId: m.Author.ID,
				code:      code,
			}

			students = append(students, st)
			sendMail(m.Author.Username, code, match[1])
			_, err := s.ChannelMessageSend(m.ChannelID, "Je t'ai envoyé un email <@"+st.discordId+"> ! :smile:")
			if err != nil {
				logrus.Error(err)
			}
		}
	}

	if strings.Contains(m.Content, "!server verify") {
		re := regexp.MustCompile("!server verify ([\\w\\d]{10})")
		match := re.FindStringSubmatch(m.Content)

		if match == nil {
			_, err := s.ChannelMessageSend(m.ChannelID, "Ce code.txt est invalide <@"+m.Author.ID+"> !")
			if err != nil {
				logrus.Error(err)
			}
		} else {

			st := Student{
				discordId: m.Author.ID,
				code:      match[1],
			}

			if studentExist(st) {
				err := s.GuildMemberRoleAdd(m.GuildID, st.discordId, "711969207381655593")
				if err != nil {
					logrus.Error(err)
					return
				}

				_, err = s.ChannelMessageSend(m.ChannelID, "SAFIR te souhaite la bienvenue <@"+st.discordId+"> ! :smile: \n\nTu peux maintenant accéder au channel <#672487866114375711> et t'assigner les rôles dont tu as besoin !\n\nN'hésite pas à ping les admins ou les modérateurs si tu as une question.")

				remove(st)

				if err != nil {
					logrus.Error(err)
					return
				}

			} else {
				_, err := s.ChannelMessageSend(m.ChannelID, "Ce code.txt est invalide <@"+m.Author.ID+"> !")
				if err != nil {
					logrus.Error(err)
				}
			}
		}
	}
}

func sendMail(pseudo string, code string, emailAddr string) {
	// Sender data.
	from := config.MailConfig.From
	password := config.MailConfig.Password

	// Receiver email address.
	to := []string{
		emailAddr,
	}

	// smtp server configuration.
	smtpHost := config.MailConfig.SMTPHost
	smtpPort := config.MailConfig.SMTPPort

	// Message.
	message := []byte("To: " + emailAddr + "\r\n" +
		"Subject: Ton code de verification\r\n\r\n" +
		"Bonjour " + pseudo + " !\n\n Voici ton code : " + code + "\n Utilise la commande \"!server verify " + code + "\" pour valider l'enregistrement.\r\n")

	// Authentication.
	auth := smtp.PlainAuth("SAFIR-URCA", from, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func studentExist(st Student) bool {
	for _, i := range students {
		if i == st {
			return true
		}
	}
	return false
}

func randSeq(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func remove(st Student) {

	for i, lst := range students {
		if lst == st {
			copy(students[i:], students[i+1:])
			students = students[:len(students)-1]
			break
		}
	}
}

func init() {

	// Load bot configuration
	if cfgFile, err := ioutil.ReadFile("config.json"); cfgFile == nil {
		log.Fatalln("Fail to read config.json", err)
	} else {
		_ = json.Unmarshal(cfgFile, &config)
	}

}

func main() {
	// Launch discord bot
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + config.DiscordConfig.Token)

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(handleDiscordMessages)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		logrus.Fatalln(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	err = dg.Close()
	if err != nil {
		logrus.Error(err)
	}
}
