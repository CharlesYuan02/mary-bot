package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
)


// mary bal
func bal(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	collectionResult, err := userCollection.FindOne(
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
	user := collectionResult.Lookup("user_name").StringValue()
	bal := collectionResult.Lookup("balance").Int32()
	return "User: " + user + "\nBalance: " + strconv.Itoa(int(bal))
}

// mary daily
func daily(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	result := userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: balance},
			}},
		},
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while inserting to database! %s\n", result.Err().Error())
		return "Error occurred while inserting to database! " + strings.Title(result.Err().Error())
	} 
	return "You have received your daily " + strconv.Itoa(balance) + " coins!"
	}

func Economy(mongoURI string, guildID int, guildName string, userID int, userName string, operation string, balance int) (string) {
	// fmt.Printf("%v %v %v %v", guildID, guildName, userID, userName)

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

	switch operation {
	case "bal":
		res := bal(ctx, userCollection, guildID, userID, balance)
		return res
	
	case "daily":
		res := daily(ctx, userCollection, guildID, userID, balance)
		return res
	
	case "insert":
		opts := options.Update().SetUpsert(true)
		collectionResult, err := userCollection.UpdateOne(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "user_id", Value: userID},
					{Key: "guild_id", Value: guildID},
					{Key: "user_name", Value: userName},
					{Key: "guild_name", Value: guildName},
					{Key: "balance", Value: balance},
				}},
			},
			opts,
		)
		if err != nil {
			fmt.Printf("Error occurred while inserting to database! %s\n", err)
			return "Error occurred while inserting to database! " + strings.Title(err.Error())
		} 
		fmt.Println(collectionResult)
	
	default: 
		return "Command not recognized!"
	}
	return "Command not recognized!"
}