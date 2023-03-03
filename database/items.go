package database

import (
	"context"
	"fmt"
	"strconv"
	"sort"
	"strings"
	"regexp" // For removing emojis
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/bwmarrin/discordgo"
)

type ShopItem struct {
    Name        string
    Price       int
    Description string
}

// Define the items for sale
var items = []ShopItem{ // Global variables don't use :=, they use =
	{"🔫 Gun", 500, "It's a gun... what do you expect?"},
	{"🚗 Car", 10000, "Run people over with this car!"},
	{"🍫 Chocolate", 50, "A great gift to give to a friend... or enemy."},
	{"💍 Ring", 1000, "Congratulations! Who's the lucky person?"},
	{"🏹 Bow", 400, "You should probably learn how to use this first..."},
}

// Lookup table for emojis
var emojiLookup = map[string]string{
	"Gun": "🔫",
	"Car": "🚗",
	"Chocolate": "🍫",
	"Ring": "💍",
	"Bow": "🏹",
}

// No return value because we are using the session to add reactions to the message
func Shop(session *discordgo.Session, message *discordgo.MessageCreate, pageSize int, currentPage int) {
	// Sort items by price
	sort.Slice(items, func(i, j int) bool {
		return items[i].Price < items[j].Price
	})

	// Check if the currentPage is out of bounds
	if currentPage < 0 {
		currentPage = 0
	} else if currentPage > len(items)/pageSize {
		currentPage = len(items) / pageSize
	}

    // Create a function to get the items for the current page
	// Make sure it displays the correct number of items and doesn't go out of bounds
    getPageItems := func() []ShopItem {
        start := currentPage * pageSize
        end := start + pageSize
        if end > len(items) {
            end = len(items)
        }
        return items[start:end]
    }

	// Create the embed
    embed := &discordgo.MessageEmbed{
        Title: "Shop",
        Color: 0xffc0cb,
        Footer: &discordgo.MessageEmbedFooter{
            Text: fmt.Sprintf("Page %d of %d", currentPage+1, len(items)/pageSize+1),
        },
    }

    // Add the items to the embed
    // Add a field for each item on the page
	pageItems := getPageItems()
    for i := range pageItems {
        item := pageItems[i]
        field := &discordgo.MessageEmbedField{
            Name: fmt.Sprintf("%s", item.Name),
			Value: fmt.Sprintf("Price: %d coins\n%s", item.Price, item.Description),
            Inline: false,
        }
        embed.Fields = append(embed.Fields, field)
    }

	// Send the embed
	_, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
	if err != nil {
		return
	}
}

func Buy(mongoURI string, guildID int, guildName string, userID int, userName string, item string, amount int) (string) {
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

	// If database for server doesn't exist, create it
	serverDatabase := client.Database(strconv.Itoa(guildID))
	userCollection := serverDatabase.Collection("Users")

	// Get user from database
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

	// Get user balance
	balance, err := collectionResult.LookupErr("balance")
	if err != nil {
		fmt.Printf("Error occurred while getting user balance! %s\n", err)
		return "Error occurred while getting user balance! " + strings.Title(err.Error())
	}

	// Get price of item specified from items
	itemPrice := 0
	for i := range items {
		// Ignore the first character of the item name because it is a unicode emoji
		// Second character is a space
		pattern := regexp.MustCompile(`^\p{So}|\s`)
		if strings.ToLower(pattern.ReplaceAllString(items[i].Name, "")) == item {
			itemPrice = items[i].Price
			break
		}
	}

	// Check if item exists
	if itemPrice == 0 {
		return "That item doesn't exist!"
	}

	// Check if user has enough money
	if balance.Int64() < int64(itemPrice) * int64(amount) {
		return "You don't have enough money to buy this item!"
	}

	// Check if the user already has this item in their inventory (in which case we +1)
	// Otherwise, we add the item to their inventory and update their balance
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
			{Key: "inventory", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{ // Check if the user has the item in their inventory
					{Key: "name", Value: item}, // If they do, update the quantity
				}},
			}},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "inventory.$.quantity", Value: amount}, // Increment the quantity by the amount specified
			}},
			{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
				{Key: "balance", Value: -itemPrice * amount}, // Decrement the balance by the price of the item * the amount specified
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s\n", err)
		return "Error occurred while updating database! " + strings.Title(err.Error())
	}

	// If the user doesn't have the item in their inventory, add it
	// This will not run if the user already has the item in their inventory because of the $not operator
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
			{Key: "inventory", Value: bson.D{
				{Key: "$not", Value: bson.D{ // Check if the user doesn't have the item in their inventory
					{Key: "$elemMatch", Value: bson.D{ 
						{Key: "name", Value: item},
					}},
				}},
			}},
		},
		bson.D{
			{Key: "$push", Value: bson.D{
				{Key: "inventory", Value: bson.D{ // Add the item to the user's inventory
					{Key: "name", Value: item}, 
					{Key: "quantity", Value: amount},
				}},
			}},
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: -itemPrice * amount}, // Decrement the balance by the price of the item * the amount specified
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s\n", err)
		return "Error occurred while updating database! " + strings.Title(err.Error())
	}

	return "You have successfully bought " + strconv.Itoa(amount) + "X " + item + " for " + strconv.Itoa(itemPrice * amount) + " coins!"
}

func Sell(mongoURI string, guildID int, guildName string, userID int, userName string, item string, amount int) (string) {
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

	// Get price of item specified from items
	itemPrice := 0
	for i := range items {
		// Ignore the first character of the item name because it is a unicode emoji
		// Second character is a space
		pattern := regexp.MustCompile(`^\p{So}|\s`)
		if strings.ToLower(pattern.ReplaceAllString(items[i].Name, "")) == item {
			itemPrice = items[i].Price
			break
		}
	}

	// Check if item exists
	if itemPrice == 0 {
		return "That item doesn't exist!"
	}

	// Check the quantity of the item the user currently has
	itemAmount := 0
	for i := range user.Inventory {
		if strings.ToLower(user.Inventory[i].Name) == item {
			itemAmount = user.Inventory[i].Quantity
			break
		}
	}

	// If the user doesn't have enough, return
	if itemAmount < amount {
		return "You don't have enough of that item to sell!"
	}

	// Set the sell price to 50% of the item's price
	itemPrice = itemPrice / 2

	// Otherwise, let the user sell the item and update their balance
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
			{Key: "inventory", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{ // Check if the user has the item in their inventory
					{Key: "name", Value: item}, // If they do, update the quantity
				}},
			}},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{ // Remember that $dec is not a thing
				{Key: "inventory.$.quantity", Value: -amount}, // Decrement the quantity by the amount specified
			}},
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: itemPrice * amount}, // Increment the balance by the price of the item * the amount specified
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating database! %s\n", err)
		return "Error occurred while updating database! " + strings.Title(err.Error())
	}

	return "You have successfully sold " + strconv.Itoa(amount) + "X " + item + " for " + strconv.Itoa(itemPrice * amount) + " coins!"
}

type User struct {
	UserID   int    `bson:"user_id"`
	GuildID  int    `bson:"guild_id"`
	Balance  int64  `bson:"balance"`
	Inventory []Item `bson:"inventory"`
}

type Item struct {
	Name     string `bson:"name"`
	Quantity int    `bson:"quantity"`
}

func Inventory(mongoURI string, guildID int, guildName string, userID int, userName string) (string, *discordgo.MessageEmbed) {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error()), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for connection is 10 secs
	defer cancel() // Fix for memory leak
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred while connecting to database! %s\n", err)
		return "Error occurred while connecting to database! " + strings.Title(err.Error()), nil
	}

	// Disconnect from database
	defer client.Disconnect(ctx) // Occurs as last line of main() function

	// Check if user exists in database
	res := IsPlaying(ctx, client, guildID, guildName, userID, userName)
	if res != "" {
		return res, nil
	}

	// Get user from database
	userCollection := client.Database(strconv.Itoa(guildID)).Collection("Users")
	filter := bson.M{"guild_id": guildID, "user_id": userID}
	var user User // User struct defined in database.go
	err = userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		fmt.Printf("Error occurred while finding user in database! %s\n", err)
		return "Error occurred while finding user in database! " + strings.Title(err.Error()), nil
	}

	// Check if user has an inventory
	if len(user.Inventory) == 0 {
		return "You do not have any items in your inventory!", nil
	}

	// Otherwise, return the user's inventory as a rich embed
	embed := &discordgo.MessageEmbed{
		Title: userName + "'s Inventory",
		Color: 0xffc0cb,
	}
	
	// Find the emoji for each item
	// Add each item to the embed
	for _, item := range user.Inventory {
		// Capitalize each item.Name
		item.Name = strings.Title(item.Name)

		// Get the emoji for the item using emojiLookup map[string]string
		emoji := emojiLookup[item.Name]
		
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%s %s", emoji, item.Name),
			Value: fmt.Sprintf("Quantity: %d", item.Quantity),
			Inline: true,
		})
	}

	return "", embed
}