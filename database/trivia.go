package database

import (
	"encoding/json"
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"github.com/bwmarrin/discordgo"
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
func Trivia(session *discordgo.Session, message *discordgo.MessageCreate) (string, *discordgo.MessageEmbed, string, string) {
	// Make a request to the trivia API
	resp, err := http.Get("https://opentdb.com/api.php?amount=1&type=multiple")
	if err != nil {
		return "Failed to get trivia question!", nil, "", ""
	}
	defer resp.Body.Close()

	// Parse the response JSON into a TriviaQuestion struct
	var triviaResponse struct {
		ResponseCode int             `json:"response_code"`
		Results      []TriviaQuestion `json:"results"`
	}
	err = json.NewDecoder(resp.Body).Decode(&triviaResponse)
	if err != nil {
		return "Failed to parse trivia question!", nil, "", ""
	}

	if triviaResponse.ResponseCode != 0 || len(triviaResponse.Results) == 0 {
		return "Failed to get trivia question!", nil, "", ""
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