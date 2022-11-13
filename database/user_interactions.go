package database

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
	commands "mary-bot/commands"
)

// mary rob @pingedUser
func rob(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, pingedUserID int) (string) {
	// Check if user is robbing themselves
	if userID == pingedUserID {
		return "You cannot rob yourself!"
	}
	
	// Check if user has enough money to rob
	userResult, err := userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()
	if err != nil {
		fmt.Printf("Error occurred while selecting from database! %s\n", err)
		return "Error occurred while selecting from database! " + strings.Title(err.Error())
	}
	userBalance := userResult.Lookup("balance").Int64()
	pingedUserName := userResult.Lookup("user_name").StringValue()
	if userBalance < 100 {
		return "That person is too poor to rob!"
	}
	
	// Check if user has robbed in the last 5 minutes
	userLastRob := userResult.Lookup("last_rob").Time()
	if time.Now().Sub(userLastRob) < 5 * time.Minute {
		return "You have already robbed someone in the last 5 minutes! Please wait " + strconv.Itoa(int(5 - time.Now().Sub(userLastRob).Minutes())) + " minutes before robbing again."
	}

	// Successful robbery
	// Generate random number between 1-50
	rand.Seed(time.Now().UnixNano())
	robAmount := rand.Intn(50) + 1
	
	// Update user's balance
	userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: robAmount},
			}},
		},
	)
	// Update pinged user's balance and user's last rob time
	userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: pingedUserID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: -robAmount},
			}},
			{Key: "$set", Value: bson.D{
				{Key: "last_rob", Value: time.Now()},
			}},
		},
	)
	return "You successfully robbed " + strconv.Itoa(robAmount) + " coins from " + pingedUserName + "!"
}

func pay(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, pingedUserID int, amount int) (string) {
	// Check if user is paying themselves 
	// Owner can pay themselves to test the command
	if userID == pingedUserID && !commands.IsOwner(userID) {
		return "You cannot pay yourself!"
	}
	
	// Check if user has enough money to pay
	userResult, err := userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()
	if err != nil {
		fmt.Printf("Error occurred while selecting from database! %s\n", err)
		return "Error occurred while selecting from database! " + strings.Title(err.Error())
	}
	userBalance := int(userResult.Lookup("balance").Int64())
	if userBalance < amount && !commands.IsOwner(userID) {
		return "You do not have enough money to pay that amount!"
	}

	// Update user's balance if not owner -> owner can pay an infinite amount
	if !commands.IsOwner(userID) {
		userCollection.FindOneAndUpdate(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "balance", Value: -amount},
				}},
			},
		)
	}

	// Update pinged user's balance
	userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: pingedUserID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: amount},
			}},
		},
	)
	// Ping user with return message
	return "You successfully paid <@" + strconv.Itoa(pingedUserID) + "> " + strconv.Itoa(amount) + " coins!"
}

// All the economy commands that require pinging another user
func UserInteraction(mongoURI string, guildID int, guildName string, userID int, userName string, pingedUserID int, operation string, amount int) (string) {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for connection is 10 secs
	defer cancel() // Fix for memory leak
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred while connecting to database! %s\n", err)
		return "Error occurred while connecting to database! " + strings.Title(err.Error())
	}

	// Disconnect from database
	defer client.Disconnect(ctx) // Occurs as last line of main() function

	// If database for server doesn't exist, create it
	serverDatabase := client.Database(strconv.Itoa(guildID))
	userCollection := serverDatabase.Collection("Users")

	// Check if user exists in database
	collectionResult, err := userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()
	_ = collectionResult // Unused variable
	if err != nil {
		// If user doesn't exist, create them
		if err == mongo.ErrNoDocuments {
			// Insert user into database
			result, err := userCollection.InsertOne(
				ctx,
				bson.D{
					{Key: "user_id", Value: userID},
					{Key: "user_name", Value: userName},
					{Key: "guild_id", Value: guildID},
					{Key: "guild_name", Value: guildName},
					{Key: "balance", Value: 0},
					{Key: "last_daily", Value: time.Now().AddDate(0, 0, -1)},
					{Key: "last_beg", Value: time.Now().AddDate(0, 0, -1)},
					{Key: "last_rob", Value: time.Now().AddDate(0, 0, -1)},
				},
			)
			if err != nil {
				fmt.Printf("Error occurred while inserting to database! %s\n", err)
				return "Error occurred while inserting to database! " + strings.Title(err.Error())
			}
			fmt.Printf("Inserted user %s into database with ID %s\n", userName, result.InsertedID)
		} else {
			fmt.Printf("Error occurred while selecting from database! %s\n", err)
			return "Error occurred while selecting from database! " + strings.Title(err.Error())
		}
	}

	// If pingedUser doesn't exist in database, send back error
	collectionResult, err = userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: pingedUserID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()
	_ = collectionResult // Unused variable
	if err != nil {
		fmt.Printf("That person is not currently playing the game!\n")
		return "That person is not currently playing the game!"
	}

	switch operation {
		case "rob":
			return rob(ctx, userCollection, guildID, userID, pingedUserID)
		case "pay":
			return pay(ctx, userCollection, guildID, userID, pingedUserID, amount)
		default: 
			return "I'm sorry, I dont recognize that command."
	}
}