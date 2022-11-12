package database

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/bson"
)

func TestConnection(mongoURI string) (string) {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for connection is 10 secs
	defer cancel() // Fix for memory leak
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred connecting to database! %s\n", err)
		return "Error occurred connecting to database! " + strings.Title(err.Error())
	}

	// Disconnect from database
	defer client.Disconnect(ctx) // Occurs as last line of main() function

	err = client.Ping(ctx, readpref.Primary()) // Pings the database
	if err != nil {
		fmt.Printf("Error occurred pinging database! %s\n", err)
		return "Error occurred pinging database! " + strings.Title(err.Error())
	}
	return ""
}

func Leaderboard(mongoURI string, guildID int) (string) {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred while connecting to database! %s\n", err)
		return "Error occurred while connecting to database! " + strings.Title(err.Error())
	}

	// Disconnect from database
	defer client.Disconnect(ctx)

	// Get leaderboard
	leaderboardCollection := client.Database(strconv.Itoa(guildID)).Collection("Users")
	leaderboardResult, err := leaderboardCollection.Find(
		ctx,
		bson.D{
			{Key: "guild_id", Value: guildID},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while selecting from database! %s\n", err)
		return "Error occurred while selecting from database! " + strings.Title(err.Error())
	}

	// Create an array for leaderboard
	var leaderboardArray []bson.M

	for leaderboardResult.Next(ctx) {
		var result bson.M
		err := leaderboardResult.Decode(&result)
		if err != nil {
			fmt.Printf("Error occurred while decoding result! %s\n", err)
			return "Error occurred while decoding result! " + strings.Title(err.Error())
		}
		// Append result to leaderboard array
		leaderboardArray = append(leaderboardArray, result)
	}

	// Sort leaderboard by balance (highest to lowest)
	sort.Slice(leaderboardArray, func(i, j int) bool {
		return leaderboardArray[i]["balance"].(int32) > leaderboardArray[j]["balance"].(int32)
	})

	// Create leaderboard string
	leaderboard := ""
	for i, result := range leaderboardArray {
		leaderboard += strconv.Itoa(i+1) + ". " + result["user_name"].(string) + ": " + strconv.Itoa(int(result["balance"].(int32))) + "\n"
	}

	return leaderboard
}