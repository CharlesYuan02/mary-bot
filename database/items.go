package database

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
	"github.com/bwmarrin/discordgo"
)

func Shop(session *discordgo.Session, message *discordgo.MessageCreate) {
    // Define the items for sale
    items := []ShopItem{
        {"item1", 100, "This is the first item"},
        {"item2", 200, "This is the second item"},
        {"item3", 300, "This is the third item"},
        {"item4", 400, "This is the fourth item"},
        {"item5", 500, "This is the fifth item"},
    }

    // Define the page size and current page
    pageSize := 2
    currentPage := 0

    // Create a function to get the items for the current page
    getPageItems := func() []ShopItem {
        start := currentPage * pageSize
        end := start + pageSize
        if end > len(items) {
            end = len(items)
        }
        return items[start:end]
    }

    // Create a function to display the current page
    displayPage := func(m *discordgo.Message) {
        // Get the items for the current page
        pageItems := getPageItems()

        // Create the embed
        embed := &discordgo.MessageEmbed{
            Title: "Shop",
            Color: 0x00ff00,
            Fields: []*discordgo.MessageEmbedField{
                {Name: "Items", Value: formatItems(pageItems), Inline: false},
            },
        }

        // Add the page information to the footer
        embed.Footer = &discordgo.MessageEmbedFooter{
            Text: fmt.Sprintf("Page %d/%d", currentPage+1, numPages(len(items), pageSize)),
        }

        // Update the message with the new embed
        _, err := session.ChannelMessageEditEmbed(m.ChannelID, m.ID, embed)
        if err != nil {
            fmt.Println(err)
        }
    }

    // Create a function to handle reactions
    reactionHandler := func(m *discordgo.MessageReactionAdd) {
        // Ignore reactions from the bot
        if m.UserID == session.State.User.ID {
            return
        }

        // Only respond to reactions on the shop message
        if m.MessageID != message.ID {
            return
        }

        // Check if the reaction is a navigation reaction
        switch m.Emoji.Name {
        case "⬅️":
            if currentPage > 0 {
                currentPage--
                displayPage(message)
            }
        case "➡️":
            if currentPage < numPages(len(items), pageSize)-1 {
                currentPage++
                displayPage(message)
            }
        }
    }

    // Send the initial shop message
    embed := &discordgo.MessageEmbed{
        Title: "Shop",
        Color: 0x00ff00,
        Fields: []*discordgo.MessageEmbedField{
            {Name: "Items", Value: formatItems(getPageItems()), Inline: false},
        },
        Footer: &discordgo.MessageEmbedFooter{
            Text: fmt.Sprintf("Page %d/%d", currentPage+1, numPages(len(items), pageSize)),
        },
    }
    m, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
    if err != nil {
        fmt.Println(err)
        return
    }

    // Add the navigation reactions to the shop message
    err = session.MessageReactionAdd(m.ChannelID, m.ID, "⬅️")
    if err != nil {
        fmt.Println(err)
    }
    err = session.MessageReactionAdd(m.ChannelID, m.ID, "➡️")
	if err != nil {
		fmt.Println(err)
	}

	// Add the reaction handler to the session
	session.AddHandler(reactionHandler)
	
	// Wait for 30 seconds
	time.Sleep(30 * time.Second)

	// Remove the reaction handler from the session
	session.RemoveHandler(reactionHandler)
}