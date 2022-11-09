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
	// Check if daily has reset
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
	lastDaily := collectionResult.Lookup("last_daily").DateTime()
	if time.Now().Unix() - lastDaily/1000 < 86400 {	
		waitTime := int(86400 - (time.Now().Unix() - lastDaily/1000))
		hours := waitTime / 3600
		minutes := (waitTime % 3600) / 60
		seconds := waitTime % 60
		return "You have already claimed your daily! Please wait " + strconv.Itoa(hours) + " hours, " + strconv.Itoa(minutes) + " minutes, and " + strconv.Itoa(seconds) + " seconds before claiming again."
	}
	
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
			{Key: "$set", Value: bson.D{
				{Key: "last_daily", Value: time.Now()},
			}},
		},
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while inserting to database! %s\n", result.Err().Error())
		return "Error occurred while inserting to database! " + strings.Title(result.Err().Error())
	} 
	return "You have received your daily " + strconv.Itoa(balance) + " coins!"
	}

func beg(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	// Check if beg has reset
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
	lastBeg := collectionResult.Lookup("last_beg").DateTime()
	// Wait one minute before begging again
	if time.Now().Unix() - lastBeg/1000 < 60 {
		waitTime := int(60 - (time.Now().Unix() - lastBeg/1000))
		return "You have already begged! Please wait " + strconv.Itoa(waitTime) + " seconds before begging again."
	}
	
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
			{Key: "$set", Value: bson.D{
				{Key: "last_beg", Value: time.Now()},
			}},
		},
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while inserting to database! %s\n", result.Err().Error())
		return "Error occurred while inserting to database! " + strings.Title(result.Err().Error())
	} 
	return "You have received " + strconv.Itoa(balance) + " coins!"
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
					{Key: "last_daily", Value: time.Now().AddDate(0, 0, -1)},
					{Key: "last_beg", Value: time.Now().AddDate(0, 0, -1)},
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
	
	case "beg":
		// Generate random value between 1 and 10
		rand.Seed(time.Now().UnixNano())
		balance := rand.Intn(10) + 1
		res := beg(ctx, userCollection, guildID, userID, balance)
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