package commands

import (
	"fmt"
	"os"
	"strconv"
	"github.com/joho/godotenv"
)

// Helper function to allow for commands by only me (the creator of the bot)
func IsOwner(userID int) (bool) {
	// Load owner user id from env vars
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading environment variables! %s\n", err)
		return false
	}

	OWNER_ID := os.Getenv("OWNER_ID")
	if OWNER_ID == "" {
		fmt.Println("Owner ID not found!")
		return false
	}

	ownerID, err2 := strconv.Atoi(OWNER_ID)
	if err2 != nil {
		fmt.Println("Owner ID is not a valid integer!")
		return false
	} else if ownerID == userID {
		return true
	}
	return false
}