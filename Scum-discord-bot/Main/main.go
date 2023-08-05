package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const (
	serverID            = "" ///BattleMetrics Server ID
	battleMetricsApiKey = ""
	discordBotToken     = ""
	discordChannelID    = ""
)

type Server struct {
	Data ServerData `json:"data"`
}

type ServerData struct {
	ID         string     `json:"id"`
	Attributes Attributes `json:"attributes"`
}
type Attributes struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	IP         string    `json:"ip"`
	Port       int       `json:"port"`
	Players    int       `json:"players"`
	MaxPlayers int       `json:"maxPlayers"`
	Rank       int       `json:"rank"`
	Location   []float64 `json:"location"`
	Status     string    `json:"status"`
	Details    struct {
		Version string `json:"version"`
		Time    string `json:"time"`
	} `json:"details"`
	Private     bool   `json:"private"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	PortQuery   int    `json:"portQuery"`
	Country     string `json:"country"`
	QueryStatus string `json:"queryStatus"`
}

var (
	StatsMessage *discordgo.Message
)

func main() {

	dg, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func updateStats(s *discordgo.Session) {
	/// Create an initial message. it will be updated after 60 seconds.
	if StatsMessage == nil {
		embed := &discordgo.MessageEmbed{
			Title: "Server Stats",
			Description: fmt.Sprintf("**Name:** %s\n**IP:** %s:%d\n**Time:** %s\n**Players:** %d/%d\n**Rank:** %d\n**Status:** %s",
				"", "0.0.0.0", 0000, "...", 0, 0, 0, "offline"),
			Color: 0x00ff00, // Green color for the embed message
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.ytimg.com/vi/NX9JnqiuEk0/hqdefault.jpg",
			},
		}
		// Send the initial stats message
		msg, err := s.ChannelMessageSendEmbed(discordChannelID, embed)
		if err != nil {
			fmt.Println("error sending message,", err)
			return
		}
		StatsMessage = msg
	} else {
		// Create an HTTP client to BattleMetrics API
		client := http.Client{}

		// Create the API request
		url := fmt.Sprintf("%s/servers/%s", "https://api.battlemetrics.com", serverID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set the API key as a request header
		req.Header.Set("Authorization", "Bearer "+battleMetricsApiKey)

		// Send the request and get the response
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
		// Create an empty buffer and use json.Indent to format the JSON response with indentation
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, body, "", "  ")
		if err != nil {
			fmt.Println("Error indenting JSON:", err)
			return
		}

		var server Server
		// Unmarshal the JSON data into the struct
		err = json.Unmarshal(body, &server)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return
		}
		///Create a embeded message.
		embed := &discordgo.MessageEmbed{
			Title: "Server Stats",
			Description: fmt.Sprintf("**Name:** %s\n**IP:** %s:%d\n**Time:** %s\n**Players:** %d/%d\n**Rank:** %d\n**Status:** %s",
				server.Data.Attributes.Name, server.Data.Attributes.IP, server.Data.Attributes.Port, server.Data.Attributes.Details.Time, server.Data.Attributes.Players, server.Data.Attributes.MaxPlayers, server.Data.Attributes.Rank, server.Data.Attributes.Status),
			Color: 0x00ff00, // Green color for the embed message
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://i.ytimg.com/vi/NX9JnqiuEk0/hqdefault.jpg",
			},
		}
		// Send the updated stats
		_, err = s.ChannelMessageEditEmbed(discordChannelID, StatsMessage.ID, embed)
		if err != nil {
			fmt.Println("error editing message,", err)
			return
		}
	}
}

// Schedule the stats update every minute
func init() {
	dg, err := discordgo.New("Bot " + discordBotToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	go func() {
		for {
			updateStats(dg)
			time.Sleep(60 * time.Second) // Sleep
		}
	}()
}
