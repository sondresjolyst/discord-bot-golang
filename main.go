package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Gopher struct {
	Name string `json: "name"`
}

var (
	Token  string
	gopher string
)

const KuteGoAPI = "http://localhost:8080"
const Prefix = "!"

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content[0:1] != Prefix {
		return
	}

	query := strings.Fields(m.Content)
	if query[0] == "!gopher" {

		// fmt.Println(len(query))
		if len(query) < 2 {
			gopher = "dr-who"
		} else {
			gopher = query[1]
		}

		// Call the KuteGo API and retrieve our cute Dr Who Gopher
		response, err := http.Get(KuteGoAPI + "/gopher/" + gopher)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			_, err = s.ChannelFileSend(m.ChannelID, "dr-who.png", response.Body)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Error: Can't get dr-who Gopher! :(", err)
		}
	}

	if m.Content == "!gophers" {
		// Call the KuteGo API and display the list of available Gophers
		response, err := http.Get(KuteGoAPI + "/gophers/")
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()

		if response.StatusCode == 200 {
			// Transform our response to a []byte
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println(err)
			}

			// Put only needed informations of the JSON document in our array of Gopher
			var data []Gopher
			err = json.Unmarshal(body, &data)
			if err != nil {
				fmt.Println(err)
			}

			// Create a string with all of the Gopher's name and a blank line as separator
			var gophers strings.Builder
			for _, gopher := range data {
				gophers.WriteString(gopher.Name + "\n")
			}

			// Send a text message with the list of Gophers
			_, err = s.ChannelMessageSend(m.ChannelID, gophers.String())
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println("Error: Can't get list of Gophers! :(")
		}
	}
}

func main() {
	// Load dotenv variables
	godotenv.Load()
	bot_token := os.Getenv("BOT_TOKEN")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + bot_token)
	if err != nil {
		fmt.Println("Error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving the messag events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discor and begin listening
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}
