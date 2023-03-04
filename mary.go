package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mary-bot/commands"
	database "mary-bot/database"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/bwmarrin/discordgo"
	// "github.com/joho/godotenv"
)

func main() {
	// Load token from env vars
	// envErr := godotenv.Load(".env")
	// if envErr != nil {
	// 	fmt.Printf("Error loading environment variables! %s\n", envErr)
	// 	return
	// }
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
	
	err := discord.Open()
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
		case strings.ToLower(command[1]) == "test" && len(command) == 2:
			session.ChannelMessageSend(message.ChannelID, "Test successful!")
		
		// mary test connection -> checks if mongoDB connection is working
		case strings.ToLower(command[1]) == "test" && strings.ToLower(command[2]) == "connection":
			dbErr := database.TestConnection(MONGO_URI)
			if dbErr != "" {
				session.ChannelMessageSend(message.ChannelID, dbErr)
			} else {
				session.ChannelMessageSend(message.ChannelID, "Database connection successful!")
			}

		// mary help -> shows all commands
		case strings.ToLower(command[1]) == "help":
			if len(command) == 2 {
				// Get Mary's avatar
				mary, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error retrieving my avatar!")
				}
				maryUser, err := mary.User("@me")
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error retrieving my avatar!")
				}
				maryAvatar := maryUser.AvatarURL("")

				// Create a rich embed
				embed := &discordgo.MessageEmbed{
					Title: "Mary's Commands",
					Color: 0xffc0cb,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: maryAvatar,
					},
					Fields: []*discordgo.MessageEmbedField{{
							Name: "mary help [optional: page number]",
							Value: "Shows all commands. The default page number is 1.",
						},{
							Name: "mary test",
							Value: "Tests if Mary is online.",
						},{
							Name: "mary test connection",
							Value: "Tests if Mary can connect to the database.",
						},{
							Name: "mary del [amount] (admin only)",
							Value: "Deletes a set number of messages.",
						},{
							Name: "mary bankrupt @user (admin only)",
							Value: "Reduces the user's balance to 0.",
						},{
							Name: "mary quote",
							Value: "Shows a random quote.",
						},{
							Name: "mary profile [optional: @user]",
							Value: "Shows your profile or a specified user's profile.",
						},{
							Name: "mary bal [optional: @user]",
							Value: "Shows your balance or a specified user's balance.",
						},{
							Name: "mary inventory",
							Value: "Shows your inventory.",
						},{ 
							Name: "mary give @user [item name] [optional: amount]",
							Value: "Gives an item to a specified user. The default amount is 1.",
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page 1/3",
					},
				}
				session.ChannelMessageSendEmbed(message.ChannelID, embed)
		
		} else if len(command) == 3 { // mary help [page number]
			// Check if the page number is a number
			pageNumber, err := strconv.Atoi(command[2])
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please enter a valid page number!")
				return
			}

			if pageNumber < 1 || pageNumber > 3 {
				session.ChannelMessageSend(message.ChannelID, "Please enter a valid page number!")
				return
			}

			// Get Mary's avatar
			mary, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving my avatar!")
			}
			maryUser, err := mary.User("@me")
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving my avatar!")
			}
			maryAvatar := maryUser.AvatarURL("")

			if pageNumber == 1 {
				// Create a rich embed
				embed := &discordgo.MessageEmbed{
					Title: "Mary's Commands",
					Color: 0xffc0cb,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: maryAvatar,
					},
					Fields: []*discordgo.MessageEmbedField{{
							Name: "mary help [optional: page number]",
							Value: "Shows all commands. The default page number is 1.",
						},{
							Name: "mary test",
							Value: "Tests if Mary is online.",
						},{
							Name: "mary test connection",
							Value: "Tests if Mary can connect to the database.",
						},{
							Name: "mary del [amount] (admin only)",
							Value: "Deletes a set number of messages.",
						},{
							Name: "mary bankrupt @user (admin only)",
							Value: "Reduces the user's balance to 0.",
						},{
							Name: "mary quote",
							Value: "Shows a random quote.",
						},{
							Name: "mary profile [optional: @user]",
							Value: "Shows your profile or a specified user's profile.",
						},{
							Name: "mary bal [optional: @user]",
							Value: "Shows your balance or a specified user's balance.",
						},{
							Name: "mary inventory",
							Value: "Shows your inventory.",
						},{
							Name: "mary give @user [item name] [optional: amount]",
							Value: "Gives an item to a specified user. The default amount is 1.",
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page 1/3",
					},
				}
				session.ChannelMessageSendEmbed(message.ChannelID, embed)
			
			} else if pageNumber == 2 {
				// Create a rich embed
				embed := &discordgo.MessageEmbed{
					Title: "Mary's Commands",
					Color: 0xffc0cb,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: maryAvatar,
					},
					Fields: []*discordgo.MessageEmbedField{{
							Name: "mary shop [optional: page number]",
							Value: "Shows the shop. You can also specify a page number.",
						},{
							Name: "mary buy [item name] [optional: amount]",
							Value: "Buys the specified item. The default amount is 1.",
						},{
							Name: "mary sell [item name] [optional: amount]",
							Value: "Sells the specified item at half the original price.",
						},{
							Name: "mary daily",
							Value: "Gives you 100 coins.",
						},{
							Name: "mary pay @user [amount]",
							Value: "Pays the mentioned user the specified amount of coins.",
						},{
							Name: "mary top/leaderboard",
							Value: "Shows the top 10 users with the highest balance.",
						},{
							Name: "mary trivia [optional: amount]",
							Value: "Starts a trivia game. Pays 50, 100, or 200 coins upon win depending on the difficulty. You can also gamble for 2X, 3X, 5X your bet.",
						},{
							Name: "mary gamble [amount]",
							Value: "Gamble the specified amount of coins.",
						},{
							Name: "mary lottery [amount]",
							Value: "Enter the lottery with 100 coins.",
						},{
							Name: "mary slots [amount]",
							Value: "Play slots with 10 coins.",
						},
					},
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page 2/3",
					},
				}
				session.ChannelMessageSendEmbed(message.ChannelID, embed)	
			} else if pageNumber == 3 {
				// Create a rich embed
				embed := &discordgo.MessageEmbed{
					Title: "Mary's Commands",
					Color: 0xffc0cb,
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: maryAvatar,
					},
					Fields: []*discordgo.MessageEmbedField{{
						Name: "mary use [item name] [@user]",
						Value: "Uses the specified item on the mentioned user. You can only use one item at a time.",
					},{
						Name: "mary eat [item name]",
						Value: "You eat a chocolate. Who knows, maybe you'll get lucky?",
					},{
						Name: "mary runover [@user]",
						Value: "Run over the mentioned user. Does not use up car item.",
					},{
						Name: "mary shoot [@user]",
						Value: "Shoot the mentioned user with the bow. Consumes one bow item.",
					},{
						Name: "mary kill [@user]",
						Value: "Shoot the mentioned user with the gun. Consumes one gun item",
					},{
						Name: "mary marry [@user]",
						Value: "Give the mentioned user a ring. If they give you one back, congratulations! You're married!",
					},{
						Name: "mary divorce [@user]",
						Value: "Divorce the mentioned user. You must be married to them or have proposed to them. Gives you back one ring.",
					},
				},
				Footer: &discordgo.MessageEmbedFooter{
					Text: "Page 3/3",
				},
				}
				session.ChannelMessageSendEmbed(message.ChannelID, embed)
			}
		}
		
		// mary profile -> shows your profile
		case strings.ToLower(command[1]) == "profile":
			// Declare variables so that they can be used outside of the if statement
			// Because apparently declaring variables inside an if statement limits their scope to the conditional
			var user string
			var bal int64
			var serverName string
			var timeLeft int
			var avatarURL string
			var spouse string

			// If user mentions another user, get their profile
			if len(message.Mentions) > 0 {
				// Get mentioned user's ID and username
				mentionedUser := message.Mentions[0]
				mentionedUserID, _:= strconv.Atoi(mentionedUser.ID)
				mentionedUserName := mentionedUser.Username

				// Get mentioned user's profile
				user, bal, serverName, timeLeft, spouse = database.GetProfile(MONGO_URI, guildID, guildName, mentionedUserID, mentionedUserName)

				// If the user variable returns the string "That person is not currently playing the game!"
				// Then return an error message
				if user == "That person is not currently playing the game!" {
					session.ChannelMessageSend(message.ChannelID, "That person is not currently playing the game!")
					time.Sleep(1 * time.Second)
					session.ChannelMessageSend(message.ChannelID, "I will add that user to the database now...")
					return
				}

				// Get mentioned user's profile picture URL
				avatarURL = mentionedUser.AvatarURL("")		
			} else {
				// Otherwise, get user's profile
				// Get username 
				userName := message.Author.Username
				// Get user's profile
				user, bal, serverName, timeLeft, spouse = database.GetProfile(MONGO_URI, guildID, guildName, userID, userName)

				if user == "That person is not currently playing the game!" {
					session.ChannelMessageSend(message.ChannelID, "You are not currently playing the game!")
					time.Sleep(1 * time.Second)
					session.ChannelMessageSend(message.ChannelID, "I will add you to the database now...")
					return
				}
				
				// Get user's profile picture URL
				avatarURL = message.Author.AvatarURL("")
			}
			
			// Extract hours, minutes and seconds from hoursUntilNextDaily
			hoursLeft := int(timeLeft)
			minutesLeft := int(hoursLeft % 60)
			secondsLeft := int(minutesLeft % 60)

			// Create embed
			embed := &discordgo.MessageEmbed{
				Title: "Profile",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: avatarURL,
				},
				Color: 0xffc0cb,
				Fields: []*discordgo.MessageEmbedField{
					{
						Name: "Username",
						Value: user,
						Inline: true,
					},
					{
						Name: "Balance",
						Value: strconv.FormatInt(bal, 10) + " coins",
						Inline: true,
					},
					{
						Name: "Server",
						Value: serverName,
						Inline: true,
					},
					{
						Name: "Married To",
						Value: spouse,
						Inline: true,
					},
					{
						Name: "Next Daily",
						Value: strconv.Itoa(hoursLeft) + "h " + strconv.Itoa(minutesLeft) + "m " + strconv.Itoa(secondsLeft) + "s",
						Inline: true,
					},
				},
			}
			// Send embed
			session.ChannelMessageSendEmbed(message.ChannelID, embed)

		// mary del (admin only) -> deletes a set number of messages
		case strings.ToLower(command[1]) == "del" && len(command) == 3:
			// Check if third argument is an integer
			_, err := strconv.Atoi(command[2])
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please enter a valid number!")
			} else {
				// Amount to delete
				amount, err := strconv.Atoi(command[2])
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while converting amount!" + strings.Title(err.Error()))
				}
				
				// Get user ID
				userID, err := strconv.Atoi(message.Author.ID)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while getting user ID!" + strings.Title(err.Error()))
				}
				
				res := commands.DeleteMessages(session, message, userID, amount)
				session.ChannelMessageSend(message.ChannelID, res)
			}
		
		// mary bankrupt (admin only) -> reduces the user's balance to 0 
		case strings.ToLower(command[1]) == "bankrupt":
			if len(command) == 3 {
				pingedUserID := strings.Trim(command[2], "<@!>")
				pingedUser, err := strconv.Atoi(pingedUserID)
				if err != nil {
					fmt.Printf("Error converting pinged user ID! %s\n", err)
				}
				res := commands.Bankrupt(MONGO_URI, guildID, userID, pingedUser)
				session.ChannelMessageSend(message.ChannelID, res)
			} else {
				session.ChannelMessageSend(message.ChannelID, "Please mention a user! Are you trying to bankrupt yourself?")
			}
		
		// mary quote -> shows a random quote
		case strings.ToLower(command[1]) == "quote":
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
		case strings.ToLower(command[1]) == "bal":
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
		
		// mary inventory -> shows user's inventory
		case strings.ToLower(command[1]) == "inventory" || strings.ToLower(command[1]) == "inv": {
			err, res := database.Inventory(MONGO_URI, guildID, guildName, userID, userName)
			if err != "" {
				session.ChannelMessageSend(message.ChannelID, err)
				break
			}
			session.ChannelMessageSendEmbed(message.ChannelID, res)
		}

		case strings.ToLower(command[1]) == "give":
			if len(command) < 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify a user to give the item to!")
				break
			} else if len(command) < 4 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an item to give!")
				break
			} 
			words := strings.Fields(message.Content)
			lastWord := words[len(words)-1] // Check if amount specified is an integer (should be last argument)
			if num, err := strconv.Atoi(lastWord); err == nil {
				// If the last word is an integer, assume the user wants to buy that many of the item
				amount := num
				item := strings.Join(command[3:len(command)-1], " ")
				pingedUser := strings.Trim(command[2], "<@!>")
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while getting user ID!" + strings.Title(err.Error()))
				}
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't give yourself an item!")
					break
				}
				res := database.Give(MONGO_URI, guildID, guildName, userID, userName, item, amount, pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			} else {
				// Assume amount to give is 1
				amount := 1
				item := strings.Join(command[3:], " ")
				pingedUser := strings.Trim(command[2], "<@!>")
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while getting user ID!" + strings.Title(err.Error()))
				}
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't give yourself an item!")
					break
				}
				res := database.Give(MONGO_URI, guildID, guildName, userID, userName, item, amount, pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			}

		// mary shop -> shows shop
		case strings.ToLower(command[1]) == "shop":
			pageSize := 3

			// If the user does not declare a page number, default to page 1 (0)
			if len(command) == 2 {
				database.Shop(session, message, pageSize, 0)
			} else if len(command) == 3 {
				// Check if third argument is an integer
				_, err := strconv.Atoi(command[2])
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please enter a valid number!")
				} else {
					// Get page number
					page, err := strconv.Atoi(command[2])
					if err != nil {
						session.ChannelMessageSend(message.ChannelID, "Error occurred while converting page number!" + strings.Title(err.Error()))
					}
					database.Shop(session, message, pageSize, page-1)
				}
			}
		
		// mary buy -> buys an item from the shop
		case strings.ToLower(command[1]) == "buy":
			// Check if user specified an item
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an item to buy!")
				break
			}
			words := strings.Fields(message.Content)
			lastWord := words[len(words)-1] // Check if amount specified is an integer (should be last argument)
			if num, err := strconv.Atoi(lastWord); err == nil {
				// If the last word is an integer, assume the user wants to buy that many of the item
				// Get item name
				item := strings.Join(command[2:len(command)-1], " ") // 0 is mary, 1 is buy
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while converting item number!" + strings.Title(err.Error()))
				}
				res := database.Buy(MONGO_URI, guildID, guildName, userID, userName, strings.ToLower(item), num)
				session.ChannelMessageSend(message.ChannelID, res)
			} else {
				// Get item name -> assume user wants to buy 1 of the item and the rest of the command is the item name
				item := strings.Join(command[2:], " ")
				res := database.Buy(MONGO_URI, guildID, guildName, userID, userName, strings.ToLower(item), 1)
				session.ChannelMessageSend(message.ChannelID, res)
			}

		// mary sell -> sells an item from the user's inventory
		case strings.ToLower(command[1]) == "sell":
			// Check if user specified an item
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an item to sell!")
				break
			}
			words := strings.Fields(message.Content)
			lastWord := words[len(words)-1] // Check if amount specified is an integer (should be last argument)
			if num, err := strconv.Atoi(lastWord); err == nil {
				// If the last word is an integer, assume the user wants to sell that many of the item
				// Get item name
				item := strings.Join(command[2:len(command)-1], " ") // 0 is mary, 1 is sell
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error occurred while converting item number!" + strings.Title(err.Error()))
				}
				res := database.Sell(MONGO_URI, guildID, guildName, userID, userName, strings.ToLower(item), num)
				session.ChannelMessageSend(message.ChannelID, res)
			} else {
				// Get item name -> assume user wants to sell 1 of the item and the rest of the command is the item name
				item := strings.Join(command[2:], " ")
				res := database.Sell(MONGO_URI, guildID, guildName, userID, userName, strings.ToLower(item), 1)
				session.ChannelMessageSend(message.ChannelID, res)
			}

		// mary daily -> gives user 100 coins
		case strings.ToLower(command[1]) == "daily":
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "daily", 100)
			session.ChannelMessageSend(message.ChannelID, res)
		
		// mary beg -> gives user 1-10 coins
		case strings.ToLower(command[1]) == "beg":
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, "beg", 0)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary rob @user -> steals 1-50 coins from user
		case strings.ToLower(command[1]) == "rob":
			pingedUserID := strings.Trim(command[2], "<@!>")
			pingedUser, err := strconv.Atoi(pingedUserID)
			if err != nil {
				fmt.Printf("Error converting pinged user ID! %s\n", err)
			}
			res := database.UserInteraction(MONGO_URI, guildID, guildName, userID, userName, pingedUser, "rob", 0)
			session.ChannelMessageSend(message.ChannelID, res)

		// mary pay @user amount -> gives user amount of coins
		case strings.ToLower(command[1]) == "pay":
			if len(command) == 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an amount to be paid!")
				return
			} else if len(command) == 4 && valid.IsInt(command[3]) == false { // &^ is bitwise AND NOT
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid amount to be paid!")
				return
			} else if strings.HasPrefix(command[3], "-") {
				session.ChannelMessageSend(message.ChannelID, "Please specify a positive amount to be paid!")
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
		case strings.ToLower(command[1]) == "leaderboard" || strings.ToLower(command[1]) == "top":
			err, res := database.Leaderboard(MONGO_URI, guildID)
			if err != "" { // Different error than usual
				session.ChannelMessageSend(message.ChannelID, err)
			}
				
			// Get profile picture of the server
			mary, err3 := discordgo.New("Bot " + os.Getenv("TOKEN"))
			if err3 != nil {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving server profile picture!")
			}
			guild, err4 := mary.Guild(message.Message.GuildID)
			if err4 != nil {
				session.ChannelMessageSend(message.ChannelID, "Error retrieving server profile picture!")
			}
			guildIconURL := guild.IconURL()

			// Create rich embed 
			embed := &discordgo.MessageEmbed{
				Title: "Leaderboard",
				Color: 0xffc0cb,
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: guildIconURL,
				},
			}

			// Use the values from res dict to create fields
			for _, data := range res {
				fields := []*discordgo.MessageEmbedField{
					{
						Name: "Rank",
						Value: strconv.Itoa(int(data["Rank"].(int))),
						Inline: true,
					},{
						Name: "Name",
						Value: data["Name"].(string),
						Inline: true,
					},{
						Name: "Balance",
						Value: strconv.FormatInt(data["Balance"].(int64), 10),
						Inline: true,
					},
				}
				embed.Fields = append(embed.Fields, fields...)
			}
			session.ChannelMessageSendEmbed(message.ChannelID, embed)

		// mary trivia -> starts a trivia game
		case strings.ToLower(command[1]) == "trivia" || strings.ToLower(command[1]) == "triv" || strings.ToLower(command[1]) == "quiz":
			gambleAmount := 0
			if len(command) == 3 {
				// Check if user specified a valid amount to gamble
				if valid.IsInt(command[2]) == false {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid amount to gamble!")
					return
				} else {
					res, err := strconv.Atoi(command[2])
					if err != nil {
						fmt.Printf("Error converting gamble amount! %s\n", err)
					}
					gambleAmount = res
					session.ChannelMessageSend(message.ChannelID, "Gambling " + command[2] + " coins. Checking balance...")
					time.Sleep(1 * time.Second)
				}
			}
			// Check if user has enough coins to gamble
			// The reason we check it here is so that if the user hasn't been added to the database yet, they will be added
			res1 := database.CheckBalance(session, message, MONGO_URI, guildID, guildName, userID, userName, gambleAmount)
			if res1 != "" {
				session.ChannelMessageSend(message.ChannelID, res1)
				return
			}

			err, res, correctAnswer, difficulty := database.Trivia(session, message, MONGO_URI, guildID, guildName, userID, userName)
			if err != "" {
				session.ChannelMessageSend(message.ChannelID, err)
			} else {
				session.ChannelMessageSendEmbed(message.ChannelID, res)

				// Wait for user to respond
				msg, err := database.WaitForResponse(session, message.ChannelID, message.Author.ID)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Error waiting for response!")
				}
				if msg == "You ran out of time!" {
					session.ChannelMessageSend(message.ChannelID, msg)
					return
				}

				// Check if user's response is correct
				if strings.ToLower(msg) == strings.ToLower(correctAnswer) {
					session.ChannelMessageSend(message.ChannelID, "Correct!")
					// Give user coins based on difficulty
					// If the user gambled coins, pay them differently 
					res := ""
					if gambleAmount != 0 {
						res = database.PayForCorrectAnswer(session, message, difficulty, MONGO_URI, guildID, guildName, userID, userName, gambleAmount)
					} else {
						res = database.PayForCorrectAnswer(session, message, difficulty, MONGO_URI, guildID, guildName, userID, userName, 0)
					}
					session.ChannelMessageSend(message.ChannelID, res)
				} else {
					session.ChannelMessageSend(message.ChannelID, "Incorrect! The correct answer is " + correctAnswer + ".")
					// If the user gambled coins, take them away
					if gambleAmount != 0 {
						session.ChannelMessageSend(message.ChannelID, "<@" + strconv.Itoa(userID) + ">, you lose. -" + command[2] + " coins.")
				}
			}
		}

		// mary gamble amount -> gamble amount of coins
		case strings.ToLower(command[1]) == "gamble":
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an amount to be gambled!")
				return
			} else if len(command) == 3 && valid.IsInt(command[2]) == false { // &^ is bitwise AND NOT
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid amount to be gambled!")
				return
			} else if strings.HasPrefix(command[2], "-") {
				session.ChannelMessageSend(message.ChannelID, "Please specify a positive amount to be gambled!")
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
		case strings.ToLower(command[1]) == "lottery":
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
		case strings.ToLower(command[1]) == "slots":
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
		
		// mary use -> uses an item from the user's inventory on a target
		case strings.ToLower(command[1]) == "use":
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an item to use!")
				break
			}
			// Get item specified
			words := strings.Fields(message.Content)
			item := strings.ToLower(words[2])
			switch item {
			case "chocolate": { // mary use chocolate 
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "chocolate", 0)
				session.ChannelMessageSend(message.ChannelID, res)
			}
			case "car": { // mary use car @target
				// Check if the user has specified a target
				if len(words) < 4 {
					session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
					break
				}
				pingedUser := strings.Trim(command[len(words)-1], "<@!>") // Get the target
				if pingedUser == "" {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				// Make sure the user doesn't use the car on themselves
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't run yourself over!")
					break
				}
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "car", pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			}
			case "gun": { // mary use gun @target
				// Check if the user has specified a target
				if len(words) < 4 {
					session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
					break
				}
				pingedUser := strings.Trim(command[len(words)-1], "<@!>") // Get the target
				if pingedUser == "" {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				// Make sure the user doesn't use the gun on themselves
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't rob yourself!")
					break
				}
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "gun", pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			} 
			case "bow": { // mary use bow @target [optional: amount]
				// Check if the user has specified a target
				if len(words) < 4 {
					session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
					break
				}
				pingedUser := strings.Trim(command[len(words)-1], "<@!>") // Get the target
				if pingedUser == "" {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				// Make sure the user doesn't use the bow on themselves
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't rob yourself!")
					break
				}
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "bow", pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			}
			case "ring": { // mary use ring @target
				// Check if the user has specified a target
				if len(words) < 4 {
					session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
					break
				}
				pingedUser := strings.Trim(command[len(words)-1], "<@!>") // Get the target
				if pingedUser == "" {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't marry yourself!")
					break
				}
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "ring", pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
			}
		}

		case strings.ToLower(command[1]) == "eat": {
			if len(command) == 2 {
				session.ChannelMessageSend(message.ChannelID, "Please specify an item to eat!")
				break
			}
			// Get item specified
			words := strings.Fields(message.Content)
			item := strings.ToLower(words[2])
			switch item {
			case "chocolate": { // mary eat chocolate
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "chocolate", 0)
				session.ChannelMessageSend(message.ChannelID, res)
				}
			default: {
				session.ChannelMessageSend(message.ChannelID, "You can't eat that!")
				}
			}
		}

		case (strings.ToLower(command[1]) == "run" && strings.ToLower(command[2]) == "over") || (strings.ToLower(command[1]) == "runover"): {
			// Check if the user has specified a target
			pingedUser := strings.Trim(command[len(command)-1], "<@!>") // Get the target
				if pingedUser == "" {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				pingedUserID, err := strconv.Atoi(pingedUser)
				if err != nil {
					session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
					break
				}
				// Make sure the user doesn't run themselves over
				if pingedUserID == userID {
					session.ChannelMessageSend(message.ChannelID, "You can't rob yourself!")
					break
				}
				res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "car", pingedUserID)
				session.ChannelMessageSend(message.ChannelID, res)
		}

		// Add in the synonyms for the specific use commands here
		case strings.ToLower(command[1]) == "shoot": {
			// Check if the user has specified a target
			if len(command) < 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
				break
			}
			pingedUser := strings.Trim(command[2], "<@!>") // Get the target
			if pingedUser == "" {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			pingedUserID, err := strconv.Atoi(pingedUser)
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			// Make sure the user doesn't use the bow on themselves
			if pingedUserID == userID {
				session.ChannelMessageSend(message.ChannelID, "You can't rob yourself!")
				break
			}
			res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "bow", pingedUserID)
			session.ChannelMessageSend(message.ChannelID, res)
		}

		case strings.ToLower(command[1]) == "kill": {
			// Check if the user has specified a target
			if len(command) < 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
				break
			}
			pingedUser := strings.Trim(command[2], "<@!>") // Get the target
			if pingedUser == "" {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			pingedUserID, err := strconv.Atoi(pingedUser)
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			// Make sure the user doesn't use the gun on themselves
			if pingedUserID == userID {
				session.ChannelMessageSend(message.ChannelID, "You can't rob yourself!")
				break
			}
			res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "gun", pingedUserID)
			session.ChannelMessageSend(message.ChannelID, res)
		}

		case strings.ToLower(command[1]) == "marry": {
			// Check if the user has specified a target
			if len(command) < 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
				break
			}
			pingedUser := strings.Trim(command[2], "<@!>")
			if pingedUser == "" {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			pingedUserID, err := strconv.Atoi(pingedUser)
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			if pingedUserID == userID {
				session.ChannelMessageSend(message.ChannelID, "You can't marry yourself!")
				break
			}
			res := database.Use(MONGO_URI, guildID, guildName, userID, userName, "ring", pingedUserID)
			session.ChannelMessageSend(message.ChannelID, res)
		}

		case strings.ToLower(command[1]) == "divorce": {
			// Check if the user has specified a target
			if len(command) < 3 {
				session.ChannelMessageSend(message.ChannelID, "Please specify a target!")
				break
			}
			pingedUser := strings.Trim(command[2], "<@!>")
			if pingedUser == "" {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			pingedUserID, err := strconv.Atoi(pingedUser)
			if err != nil {
				session.ChannelMessageSend(message.ChannelID, "Please specify a valid target!")
				break
			}
			if pingedUserID == userID {
				session.ChannelMessageSend(message.ChannelID, "You can't marry yourself!")
				break
			}
			res := database.Divorce(MONGO_URI, guildID, guildName, userID, userName, pingedUserID)
			session.ChannelMessageSend(message.ChannelID, res)
		}

		// Everything else (will most likely return "I'm sorry, I dont recognize that command.")
		default:
			res := database.Economy(MONGO_URI, guildID, guildName, userID, userName, command[1], 0)
			session.ChannelMessageSend(message.ChannelID, res)
		}
	}
}