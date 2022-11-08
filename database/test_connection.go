package database

import (
	"context"
	"fmt"
	"strings"
	"time"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func TestConnection(mongoURI string) (string) {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Printf("Error occurred creating MongoDB client! %s\n", err)
		return "Error occurred creating MongoDB client! " + strings.Title(err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second) // Timeout for connection is 10 secs
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