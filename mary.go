package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"github.com/joho/godotenv"
	"github.com/bwmarrin/discordgo"
	valid "github.com/asaskevich/govalidator"
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

	// Set Mary's status (make sure to do this after discord.Open())
	err = discord.UpdateGameStatus(0, "with her sister Eve")
	if err != nil {
		fmt.Printf("Error setting Mary's status! %s\n", err)
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
		} else {
			return
		}
	}

	// Get URI for connecting to MongoDB database
	MONGO_URI := os.Getenv("MONGO_URI")
	if MONGO_URI == "" {
		fmt.Println("MongoDB URI not found!")
		return
	}

	// Get guild ID and name
	guild, err1 := session.Guild(message.GuildID)
	guildID, err2 := strconv.Atoi(guild.ID)
	guildName := guild.Name
	if err1 != nil {
		fmt.Printf("Error retrieving guild details! %s\n", err1)
	} else if err2 != nil {
		fmt.Printf("Error converting guild ID! %s\n", err1)
	}
	userID, err3 := strconv.Atoi(message.Author.ID)
	if err3 != nil {
		fmt.Printf("Error converting user ID! %s\n", err1)
	}
	userName := message.Author.Username

	command := strings.Split(message.Content, " ")
	if strings.ToLower(command[0]) == "mary" {
		switch true {
		
		// mary test
		case command[1] == "test" && len(command) == 2:
			session.ChannelMessageSend(message.ChannelID, "Test successful!")
		
		// mary test connection -> checks if mongoDB connection is working
		case command[1] == "test" && command[2] == "connection":
			dbErr := database.TestConnection(MONGO_URI)
			if dbErr != "" {
				session.ChannelMessageSend(message.ChannelID, dbErr)
			} else {
				session.ChannelMessageSend(message.ChannelID, "Database connection successful!")
			}

		// mary help -> shows all commands
		case command[1] == "help":
			session.ChannelMessageSend(message.ChannelID, "```Commands:\n\n" +
				"mary help -> shows all commands\n" +
				"mary test -> tests if Mary is online\n" +
				"mary test connection -> tests if Mary can connect to the database\n" +
				"mary quote -> shows a random quote\n" +
				"mary bal -> shows your balance\n" +
				"mary bal @user -> shows the balance of the mentioned user\n" +
				"mary daily -> gives you 100 coins\n" +
				"mary pay/give @user [amount] -> pays the mentioned user the specified amount of coins\n" +
				"mary top/leaderboard -> shows the top 10 users with the most coins\n" +
				"mary gamble [amount] -> gamble the specified amount of coins\n" +
				"mary lottery [amount] -> enter the lottery with 100 coins\n" +
				"mary slots [amount] -> play slots with 10 coins\n" +
				"```")
		
		// mary quote -> shows a random quote
		case command[1] == "quote":
			quote, err := http.Get("https://api.quotable.io/random")
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving quote!")
			} else {
				defer quote.Body.Close()
				quoteData, err := ioutil.ReadAll(quote.Body)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error retrieving quote!")
				} else {
					var quoteJSON map[string]interface{}
					json.Unmarshal(quoteData, &quoteJSON)
					session.ChannelMessageSend(message.ChannelID, fmt.Sprintf("```%s\n\n- %s```", quoteJSON["content"], quoteJSON["author"]))
				}
			}
		
		// mary bal -> checks balance of message author
		case command[1] == "bal":
			// Return balance of user
			if len(command) == 2 {
				res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "bal", 0)
				session.ChannelMessageSend(message.ChannelID, res)
			} else if len(command) == 3 {
				// Return balance of mentioned user
				if strings.HasPrefix(command[2], "<@") && strings.HasSuffix(command[2], ">") {
					mentionedUser := strings.TrimPrefix(command[2], "<@")
					mentionedUser = strings.TrimSuffix(mentionedUser, ">")
					mentionedUser = strings.TrimPrefix(mentionedUser, "!")
					mentionedUserID, err := strconv.Atoi(mentionedUser)
					if err != nil {
						session.ChannelMessageSend(message.ChannelID, "Error retrieving balance!")
					} else {
						res := database.Economy(MONGO_URI, guildID, guildName, mentionedUserID, "", "bal", 0)
						session.ChannelMessageSend(message.ChannelID, res)
					}
				} else {
					session.ChannelMessageSend(message.ChannelID, "Error retrieving balance!")
				}
			} else {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving balance!")
			}

		// mary daily -> gives user 100 coins
		case command[1] == "daily":
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "daily", 100)
			session.ChannelMessageSend(message.ChannelID, res)
		
		// mary beg -> gives user 1-10 coins
		case command[1] == "beg":
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "beg", 0)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary rob @user -> steals 1-50 coins from user
		case command[1] == "rob":
			pingedUserID := strings.Trim(command[2], "<@!>")
			pingedUser, err := strconv.Atoi(pingedUserID)
			if err != nil {
				fmt.Printf("Error converting pinged user ID! %s\n", err)
			}
			res := database.UserInteraction(MONGO_URI, guildID, guildName, userID, userName, pingedUser, "rob", 0)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary pay/give @user amount -> gives user amount of coins
		case command[1] == "pay" || command[1] == "give":
			if len(command) == 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an amount to be paid!")
				return
			} else if len(command) == 4 && valid.IsInt(command[3]) == false { // &^ is bitwise AND NOT
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid amount to be paid!")
				return
			}
			pingedUserID := strings.Trim(command[2], "<@!>")
			pingedUser, err := strconv.Atoi(pingedUserID)
			if err != nil {
				fmt.Printf("Error converting pinged user ID! %s\n", err)
			}
			amount, err := strconv.Atoi(command[3])
			if err != nil {
				fmt.Printf("Error converting amount! %s\n", err)
			}
			res := database.UserInteraction(MONGO_URI, guildID, guildName, userID, userName, pingedUser, "pay", amount)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary top/leaderboard -> shows users with the most coins in descending order
		case command[1] == "leaderboard" || command[1] == "top":
			res := database.Leaderboard(MONGO_URI, guildID)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary gamble amount -> gamble amount of coins
		case command[1] == "gamble":
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an amount to be gambled!")
				return
			} else if len(command) == 3 && valid.IsInt(command[2]) == false { // &^ is bitwise AND NOT
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid amount to be gambled!")
				return
			} else {
				session.ChannelMessageSend(message.ChannelID, "Gambling " + command[2] + " coins...")
				time.Sleep(1 * time.Second)
			}
			amount, err := strconv.Atoi(command[2])	
			if err != nil {
				fmt.Printf("Error converting amount! %s\n", err)
			}
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "gamble", amount)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary lottery -> enter lottery for 100 coins
		case command[1] == "lottery":
			if len(command) > 2 {
				session.ChannelMessageSend(message.ChannelID, "You can only spend 100 coins on the lottery!")
				time.Sleep(500 * time.Millisecond)
				session.ChannelMessageSend(message.ChannelID, "Gambling 100 coins...")
				time.Sleep(1 * time.Second)
			} else {
				session.ChannelMessageSend(message.ChannelID, "Gambling 100 coins...")
				time.Sleep(1 * time.Second)
			}
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "lottery", 100)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary slots -> play slots for 10 coins
		case command[1] == "slots":
			if len(command) > 2 {
				session.ChannelMessageSend(message.ChannelID, "You can only spend 10 coins on slots!")
				time.Sleep(500 * time.Millisecond)
				session.ChannelMessageSend(message.ChannelID, "Gambling 10 coins...")
				time.Sleep(1 * time.Second)
			} else {
				session.ChannelMessageSend(message.ChannelID, "Gambling 10 coins...")
				time.Sleep(1 * time.Second)
			}
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "slots", 10)
			session.ChannelMessageSend(message.ChannelID, res)

		// Everything else (will most likely return "I'm sorry, I dont recognize that command.")
		default:
			fmt.Printf("%T\n", command[1])
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, command[1], 0)
			session.ChannelMessageSend(message.ChannelID, res)
		}
	}
}