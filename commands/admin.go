package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"github.com/bwmarrin/discordgo"
	//"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
)

// Helper function to allow for commands by only me (the creator of the bot)
func IsOwner(userID int) (bool) {
	// Load owner user id from env vars
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	fmt.Printf("Error loading environment variables! %s\n", err)
	// 	return false
	// }

	OWNER_ID := os.Getenv("OWNER_ID")
	if OWNER_ID == "" {
		fmt.Println("Owner ID not found!")
		return false
	}

	ownerID, err2 := strconv.Atoi(OWNER_ID)
	if err2 != nil {
		fmt.Println("Owner ID is not a valid integer!")
		return false
	} else if ownerID == userID {
		return true
	}
	return false
}

func DeleteMessages(session *discordgo.Session, message *discordgo.MessageCreate, userID int, amount int) (string) {
	// Check if user is owner
	if !IsOwner(userID) {
		return "Apologies, this command is not available to you."
	}

	// Get the previous amount of messages 
	// Delete amount messages (not including the command message)
	messages, err := session.ChannelMessages(message.ChannelID, amount+1, "", "", "")
	if err != nil {
		return "Error occurred while getting messages! " + strings.Title(err.Error())
	}

	// Delete the messages
	for _, message := range messages {
		err = session.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			return "Error occurred while deleting messages! " + strings.Title(err.Error())
		}
	}
	return "Successfully deleted " + strconv.Itoa(amount) + " messages!"
}

func Bankrupt(mongoURI string, guildID int, userID int, pingedUserID int) (string) {
	// Check if user is owner
	if !IsOwner(userID) {
		return "Apologies, this command is not available to you."
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error())
	}
	ctx := context.Background()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred connecting to MongoDB! %s\n", err)
		return "Error occurred connecting to MongoDB! " + strings.Title(err.Error())
	}
	defer client.Disconnect(ctx)

	// Get the user collection
	collection := client.Database(strconv.Itoa(guildID)).Collection("Users")

	// Update the balance of the pinged user to 0 and get user name
	filter := bson.D{{Key: "user_id", Value: pingedUserID}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "balance", Value: int64(0)}}}}
	var result bson.M
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&result)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s\n", err)
		return "That person is not currently playing the game!"
	}
	// ping user with <@!user_id> to get their name
	return "<@!" + strconv.Itoa(pingedUserID) + ">, you are now bankrupt!"
}