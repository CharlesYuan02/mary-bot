package database

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/bwmarrin/discordgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" 
	"go.mongodb.org/mongo-driver/bson"
	commands "mary-bot/commands"
)

// TriviaQuestion represents a single trivia question from the API
type TriviaQuestion struct {
	Category     string   `json:"category"`
	Type         string   `json:"type"`
	Difficulty   string   `json:"difficulty"`
	Question     string   `json:"question"`
	Correct      string   `json:"correct_answer"`
	Incorrect    []string `json:"incorrect_answers"`
}


// Trivia is a function that starts a trivia game session
func Trivia(session *discordgo.Session, message *discordgo.MessageCreate, mongoURI string, guildID int, guildName string, userID int, userName string) (string, *discordgo.MessageEmbed, string, string) {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error()), nil, "", ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for connection is 10 secs
	defer cancel() // Fix for memory leak
	err = client.Connect(ctx)
	if err != nil {
		fmt.Printf("Error occurred connecting to database! %s\n", err)
		return "Error occurred while connecting to database! " + strings.Title(err.Error()), nil, "", ""
	}

	// Disconnect from database
	defer client.Disconnect(ctx) // Occurs as last line of main() function

	// Get user from database
	serverDatabase := client.Database(strconv.Itoa(guildID))
	userCollection := serverDatabase.Collection("Users")
	collectionResult, err := userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()

	// Get user
	lastTrivia := collectionResult.Lookup("last_trivia").DateTime()

	// Wait 5 seconds before playing trivia again
	if time.Now().Unix() - lastTrivia/1000 < 5 && commands.IsOwner(userID) == false{
		return "<@" + strconv.Itoa(userID) + ">, you must wait 5 seconds before playing trivia again!", nil, "", ""
	}

	// If the user is not on cooldown, set the last_trivia field to now
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "last_trivia", Value: time.Now()},
			}},
		},
	)

	// Make a request to the trivia API
	resp, err := http.Get("https://opentdb.com/api.php?amount=1&type=multiple")
	if err != nil {
		fmt.Printf("Failed to get trivia question! %s\n", err)
		return "Failed to get trivia question!" + strings.Title(err.Error()), nil, "", ""
	}
	defer resp.Body.Close()

	// Parse the response JSON into a TriviaQuestion struct
	var triviaResponse struct {
		ResponseCode int             `json:"response_code"`
		Results      []TriviaQuestion `json:"results"`
	}
	err = json.NewDecoder(resp.Body).Decode(&triviaResponse)
	if err != nil {
		fmt.Printf("Failed to parse trivia question! %s\n", err)
		return "Failed to parse trivia question!" + strings.Title(err.Error()), nil, "", ""
	}

	if triviaResponse.ResponseCode != 0 || len(triviaResponse.Results) == 0 {
		fmt.Printf("Failed to get trivia question! Response code: %d\n", triviaResponse.ResponseCode)
		return "Failed to get trivia question!" + strings.Title(err.Error()), nil, "", ""
	}

	// Select the first question from the response and shuffle the answer choices
	question := triviaResponse.Results[0]
	choices := append(question.Incorrect, question.Correct)
	rand.Shuffle(len(choices), func(i, j int) {
		choices[i], choices[j] = choices[j], choices[i]
	})

	// Format the choices with A, B, C, D
	formattedChoices := []string{"A", "B", "C", "D"}
	for i, choice := range choices {
		formattedChoices[i] += ") " + choice
	}

	// Decode the HTML entities in the question and choices
	question.Question = html.UnescapeString(question.Question)
	for i, choice := range formattedChoices {
		formattedChoices[i] = html.UnescapeString(choice)
	}

	// Get the letter corresponding to the correct answer
	correctLetter := ""
	for i, choice := range choices {
		if choice == question.Correct {
			correctLetter = string(formattedChoices[i][0])
			break
		}
	}

	// Capitalize the first letter of the difficulty
	question.Difficulty = strings.Title(question.Difficulty)

	// Send the question as a rich embed
	embed := &discordgo.MessageEmbed{
		Title:       question.Question,
		Description: fmt.Sprintf("Category: %s \nDifficulty: %s", question.Category, question.Difficulty),
		Color:       0xffc0cb,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Choices", Value: strings.Join(formattedChoices, "\n"), Inline: false},
		},
	}

	// Return the embed
	return "", embed, correctLetter, question.Difficulty
}

// Wait for the user to respond; This is the equivalent of channelMessageWait in Discord.js
func WaitForResponse(session *discordgo.Session, channelID string, authorID string) (string, error) {
	// Create a channel for receiving the user's response
    responseChan := make(chan string)

    // Create a message handler that listens for the user's response
    handler := func(s *discordgo.Session, m *discordgo.MessageCreate) {
        // Ignore messages from other users or channels
        if m.Author.ID != authorID || m.ChannelID != channelID {
            return
        }

        // Send the response to the response channel
        responseChan <- m.Content
    }

    // Add the message handler to the session
    session.AddHandler(handler)

    // Wait for the user's response or for a timeout of 10 seconds
    select {
    case response := <-responseChan:
        return response, nil 
    case <-time.After(10 * time.Second):
        return "You ran out of time!", nil // Return an error indicating a timeout
    }
}

// Pay the user for their correct answer
func PayForCorrectAnswer(session *discordgo.Session, message *discordgo.MessageCreate, difficulty string, mongoURI string, guildID int, guildName string, userID int, userName string, amount int) (string) {
	// Calculate the amount of coins to pay the user
	if amount == 0 {
		switch strings.ToLower(difficulty) {
		case "easy":
			amount = 50
		case "medium":
			amount = 100
		case "hard":
			amount = 200
		}
	} else if amount > 0 {
		switch strings.ToLower(difficulty) {
		case "easy":
			amount *= 2
		case "medium":
			amount *= 3
		case "hard":
			amount *= 5
		}
	}

	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! " + strings.Title(err.Error()))
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

	// Get the correct database and collection
	serverDatabase := client.Database(strconv.Itoa(guildID))
	userCollection := serverDatabase.Collection("Users")

	// Update the user's balance
	_, err = userCollection.UpdateOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
		bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "balance", Value: amount},
			}},
		},
	)
	if err != nil {
		fmt.Printf("Error occurred while updating user's balance! %s\n", err)
		return "Error occurred while updating user's balance! " + strings.Title(err.Error())
	}
	// Success
	return "<@" + strconv.Itoa(userID) + ">, you have been paid " + strconv.Itoa(amount) + " coins!"
}

// Check if the user has enough coins to gamble
// Also check if the user is playing the game
func CheckBalance(session *discordgo.Session, message *discordgo.MessageCreate, mongoURI string, guildID int, guildName string, userID int, userName string, amount int) (string) {
	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! " + strings.Title(err.Error()))
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
	serverDatabase := client.Database(strconv.Itoa(guildID))
	userCollection := serverDatabase.Collection("Users")
	collectionResult, err := userCollection.FindOne(
		ctx,
		bson.D{
			{Key: "user_id", Value: userID},
			{Key: "guild_id", Value: guildID},
		},
	).DecodeBytes()
	
	// Find user
	collectionResult, err2 := userCollection.FindOne(context.Background(), bson.D{
		{Key: "user_id", Value: userID},
		{Key: "guild_id", Value: guildID},
	}).DecodeBytes() 
	if err2 != nil {
		fmt.Printf("Error occurred while finding user! %s\n", err2)
		return "Error occurred while finding user! " + strings.Title(err2.Error())
	}

	// Check if user has enough to gamble
	userBalance := collectionResult.Lookup("balance").Int64()
	if userBalance < int64(amount) {
		return "<@" + strconv.Itoa(userID) + ">, you don't have enough coins to gamble that much!"
	}
	// Success
	return ""
}