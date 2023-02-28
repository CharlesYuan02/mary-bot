package database

import (
	"fmt"
	//"strconv"
	//"strings"
	"sort"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options" 
	// "go.mongodb.org/mongo-driver/bson"
	"github.com/bwmarrin/discordgo"
)

type ShopItem struct {
    Name        string
    Price       int
    Description string
}

// No return value because we are using the session to add reactions to the message
func Shop(session *discordgo.Session, message *discordgo.MessageCreate, pageSize int, currentPage int) {
	// Define the items for sale
	items := []ShopItem{
        {"🔫 Gun", 500, "It's a gun... what do you expect?"},
        {"🚗 Car", 10000, "Run people over with this car!"},
        {"🍫 Chocolate", 50, "A great gift to give to a friend... or enemy."},
        {"💍 Ring", 1000, "Congratulations! Who's the lucky person?"},
        {"🏹 Bow", 400, "You should probably learn how to use this first..."},
    }

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

