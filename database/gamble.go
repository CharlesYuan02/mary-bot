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


func Gamble(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	// Subtract balance from user
	result := userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: -balance},
			}},
		},
		options.FindOneAndUpdate().SetUpsert(true),
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while updating database! %s\n", result.Err())
		return "Error occurred while updating database! " + strings.Title(result.Err().Error())
	}

	// Get last_gamble from database
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

	lastGamble := collectionResult.Lookup("last_gamble").DateTime()
	// Wait ten seconds before gambling again
	if time.Now().Unix() - lastGamble/1000 < 10 && commands.IsOwner(userID) == false {
		return "You must wait 10 seconds before gambling again!"
	}

	// Roll dice 
	dice := rand.Intn(100) + 1
	if dice <= 50 {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	} else if dice <= 80 {
		// Win - 30% chance
		result := userCollection.FindOneAndUpdate(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "balance", Value: balance * 2},
				}},
			},
			options.FindOneAndUpdate().SetUpsert(true),
		)
		if result.Err() != nil {
			fmt.Printf("Error occurred while updating database! %s\n", result.Err())
			return "Error occurred while updating database! " + strings.Title(result.Err().Error())
		}
		return "You win! +" + strconv.Itoa(balance * 2) + " coins!"
	} else {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	}
}

func Lottery(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	// Subtract balance from user
	result := userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: -balance},
			}},
		},
		options.FindOneAndUpdate().SetUpsert(true),
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while updating database! %s\n", result.Err())
		return "Error occurred while updating database! " + strings.Title(result.Err().Error())
	}

	// Get last_gamble from database
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

	lastGamble := collectionResult.Lookup("last_gamble").DateTime()
	// Wait ten seconds before gambling again
	if time.Now().Unix() - lastGamble/1000 < 10 && commands.IsOwner(userID) == false {
		return "You must wait 10 seconds before gambling again!"
	}

	// Roll dice 
	dice := rand.Intn(100) + 1
	if dice <= 60 {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	} else if dice <= 70 || dice > 90 {
		// Win - 20% chance but 5X the payout
		result := userCollection.FindOneAndUpdate(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "balance", Value: balance * 5},
				}},
			},
			options.FindOneAndUpdate().SetUpsert(true),
		)
		if result.Err() != nil {
			fmt.Printf("Error occurred while updating database! %s\n", result.Err())
			return "Error occurred while updating database! " + strings.Title(result.Err().Error())
		}
		return "You win! +" + strconv.Itoa(balance * 5) + " coins!"
	} else {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	}
}

func Slots(ctx context.Context, userCollection *mongo.Collection, guildID int, userID int, balance int) (string) {
	// Subtract balance from user
	result := userCollection.FindOneAndUpdate(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: -balance},
			}},
		},
		options.FindOneAndUpdate().SetUpsert(true),
	)
	if result.Err() != nil {
		fmt.Printf("Error occurred while updating database! %s\n", result.Err())
		return "Error occurred while updating database! " + strings.Title(result.Err().Error())
	}

	// Get last_gamble from database
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

	lastGamble := collectionResult.Lookup("last_gamble").DateTime()
	// Wait ten seconds before gambling again
	if time.Now().Unix() - lastGamble/1000 < 10 && commands.IsOwner(userID) == false {
		return "You must wait 10 seconds before gambling again!"
	}

	// Roll dice 
	dice := rand.Intn(100) + 1
	if dice <= 40 {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	} else if dice <= 70 || dice > 90 {
		// Win - 40% chance and 2X payout, but can only win 20 coins at a time
		result := userCollection.FindOneAndUpdate(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{
					{Key: "balance", Value: balance * 2},
				}},
			},
			options.FindOneAndUpdate().SetUpsert(true),
		)
		if result.Err() != nil {
			fmt.Printf("Error occurred while updating database! %s\n", result.Err())
			return "Error occurred while updating database! " + strings.Title(result.Err().Error())
		}
		return "You win! +" + strconv.Itoa(balance * 2) + " coins!"
	} else {
		// Lose
		return "You lose. -" + strconv.Itoa(balance) + " coins."
	}
}
