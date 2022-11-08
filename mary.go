package main 

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
	database "mary-bot/database"
)

func main() {
	// Load token from env vars
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading environment variables! %s\n", err)
		return
	}
	TOKEN := os.Getenv("TOKEN")
	if TOKEN == "" {
		fmt.Println("Token not found!")
		return
	}
	
	discord, discordError := discordgo.New("Bot " + TOKEN)
	if discordError != nil {
		fmt.Printf("Error creating Discord session! %s\n", discordError)
		return
	}

	// Handler for sending messages
	// Remember to go on Developer Portal, Bot and enable Privileged Gateway Intents (not enabled by default)
	// https://github.com/bwmarrin/discordgo/issues/1264
	discord.Identify.Intents = discordgo.IntentMessageContent
	discord.AddHandler(createMessage)
	discord.Identify.Intents = discordgo.IntentsGuildMessages
	
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord connection!")
		return
	}
	
	fmt.Println("Mary, online and ready!")

	sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
    <-sc
	discord.Close()
}

func createMessage(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages sent by Mary herself
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content == "test" {
		_, err := session.ChannelMessageSend(message.ChannelID, "Test successful!")
		if err != nil {
			fmt.Printf("Error occurred during testing! %s\n", err)
		}
	}

	// Get URI for connecting to MongoDB database
	MONGO_URI := os.Getenv("MONGO_URI")
	if MONGO_URI == "" {
		fmt.Println("MongoDB URI not found!")
		return
	}

	command := strings.Split(message.Content, " ")
	if command[0] == "mary" {
		switch true {
		// mary test
		case command[1] == "test" && len(command) == 2:
			session.ChannelMessageSend(message.ChannelID, "Test successful!")
			break
		// mary test connection -> checks if mongoDB connection is working
		case command[1] == "test" && command[2] == "connection":
			dbErr := database.TestConnection(MONGO_URI)
			if dbErr != "" {
				session.ChannelMessageSend(message.ChannelID, dbErr)
				break
			} else {
				session.ChannelMessageSend(message.ChannelID, "Database connection successful!")
			break
			}
		}
	}
}