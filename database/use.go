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

// Structs are defined in items.go


func Use(mongoURI string, guildID int, guildName string, userID int, userName string, item string, amount int, pingedUserID int) (string) {
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

	// Check if user exists in database
	res := IsPlaying(ctx, client, guildID, guildName, userID, userName)
	if res != "" {
		return res
	}

	// Get user from database
	userCollection := client.Database(strconv.Itoa(guildID)).Collection("Users")
	filter := bson.M{"guild_id": guildID, "user_id": userID}
	var user User // User struct defined in database.go
	err = userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Printf("Error occurred while finding user in database! %s\n", err)
		return "Error occurred while finding user in database! " + strings.Title(err.Error())
	}

	// Check if user has an inventory
	if len(user.Inventory) == 0 {
		return "You do not have any items in your inventory!"
	}

	// Check if user has the item in their inventory
	for _, i := range user.Inventory {
		if i.Name == item {
			// Check if the user has enough of the item
			if i.Quantity < amount {
				return "You do not have enough of that item in your inventory to use!"
			}
		}
	}

	// Check if the user has waited a minute since their last use indicated by last_use
	// If the user has not waited a minute, return an error
	lastUse := user.LastUse
	if time.Since(lastUse) < time.Minute {
		return "You must wait a minute between uses!"
	}
	
	// Update the user's last_use to the current time
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "last_use", Value: time.Now()},
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s", err)
		return "Error occurred while updating database! " + strings.Title(err.Error())
	}

	// Update the user's inventory to reduce the amount of the item they have
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
			{Key: "inventory.item.name", Value: item}, // Get the item in the inventory 
		},
		bson.D{
			{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
				{Key: "inventory.$.quantity", Value: -amount},
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s\n", err)
		return "Error occurred while updating database! " + strings.Title(err.Error())
	}
	
	// Check what the item is
	// The check for whether a pingedUser exists is done in mary.go
	switch item {
	case "gun": 
		// Check if the pinged user exists in the database
		pingedUserCollection := client.Database(strconv.Itoa(guildID)).Collection("Users")
		pingedUserFilter := bson.M{"guild_id": guildID, "user_id": pingedUserID}
		var pingedUser User
		err = pingedUserCollection.FindOne(ctx, pingedUserFilter).Decode(&pingedUser)
		if err != nil {
			fmt.Printf("That user is not currently playing the game!\n")
			return "That user is not currently playing the game!"
		}

		// Check if the pinged user has enough money to rob
		pingedUserBalance := pingedUser.Balance
		if pingedUserBalance < 100 {
			return "That user does not have enough money for you to rob!"
		}

		// Otherwise, get the pinged user balance and rob them for a random percentage amount
		robbedAmount := int64(float64(pingedUserBalance) * (rand.Float64() * 0.5 + 0.1)) // Random percentage between 10% and 60%
		_, err = pingedUserCollection.UpdateOne(
			ctx,
			bson.D{
				{Key: "user_id", Value: pingedUserID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
					{Key: "balance", Value: -robbedAmount}, // Decrement the balance by the robbed amount
				}},
			},
		)
		if err != nil {
			fmt.Printf("Error occurred while updating database! %s\n", err)
			return "Error occurred while updating database! " + strings.Title(err.Error())
		}

		// Update the user's balance
		_, err = userCollection.UpdateOne(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{ 
					{Key: "balance", Value: robbedAmount}, // Increment the balance by the robbed amount
				}},
			},
		)

		return "You held up <@" + strconv.Itoa(pingedUserID) + "> at gunpoint and robbed " + strconv.Itoa(int(robbedAmount)) + " coins from them!"

	case "bow":
		// Check if the pinged user exists in the database
		pingedUserCollection := client.Database(strconv.Itoa(guildID)).Collection("Users")
		pingedUserFilter := bson.M{"guild_id": guildID, "user_id": pingedUserID}
		var pingedUser User
		err = pingedUserCollection.FindOne(ctx, pingedUserFilter).Decode(&pingedUser)
		if err != nil {
			fmt.Printf("That user is not currently playing the game!\n")
			return "That user is not currently playing the game!"
		}

		// Check if the pinged user has enough money to rob
		pingedUserBalance := pingedUser.Balance
		if pingedUserBalance < 100 {
			return "That user does not have enough money for you to rob!"
		}

		// Check if the pinged user has a gun
		pingedUserInventory := pingedUser.Inventory
		hasGun := false
		for _, item := range pingedUserInventory {
			if item.Name == "gun" {
				hasGun = true
				break
			}
		}

		robbedAmount := int64(float64(pingedUserBalance) * (rand.Float64() * 0.1 + 0.2)) // Random percentage between 20% and 30%

		// If the pinged user has a gun, then they you lost a percentage of your balance
		if hasGun {
			_, err = userCollection.UpdateOne(
				ctx,
				bson.D{
					{Key: "user_id", Value: userID},
					{Key: "guild_id", Value: guildID},
				},
				bson.D{
					{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
						{Key: "balance", Value: -int64(robbedAmount)},
					}},
				},
			)
			if err != nil {
				fmt.Printf("Error occurred while updating database! %s\n", err)
				return "Error occurred while updating database! " + strings.Title(err.Error())
			}

			return "You tried to rob <@" + strconv.Itoa(pingedUserID) + "> with a bow, but they had a gun and shot you! You lost " + strconv.Itoa(int(robbedAmount)) + " coins!"
		} else {
		_, err = pingedUserCollection.UpdateOne(
			ctx,
			bson.D{
				{Key: "user_id", Value: pingedUserID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
					{Key: "balance", Value: -robbedAmount}, // Decrement the balance by the robbed amount
				}},
			},
		)
		if err != nil {
			fmt.Printf("Error occurred while updating database! %s\n", err)
			return "Error occurred while updating database! " + strings.Title(err.Error())
		}

		// Update the user's balance
		_, err = userCollection.UpdateOne(
			ctx,
			bson.D{
				{Key: "user_id", Value: userID},
				{Key: "guild_id", Value: guildID},
			},
			bson.D{
				{Key: "$inc", Value: bson.D{ 
					{Key: "balance", Value: robbedAmount}, // Increment the balance by the robbed amount
				}},
			},
		)

		return "You shot <@" + strconv.Itoa(pingedUserID) + "> and took " + strconv.Itoa(int(robbedAmount)) + " coins from them!"
	}
	}
	return "" 
}