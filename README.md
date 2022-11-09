# mary-bot
<img src="https://i.ytimg.com/vi/9S831972rjA/maxresdefault.jpg">

## Introduction
A Discord bot I created in Go. Her name is Mary. After receiving my offer to work at Uber, I knew I had to learn Go right away, as my interviewer informed me that that was the primary language they used for back end development. I decided to learn some noSQL and MongoDB to implement a database-driven game so that people will actually interact with her (cuz my classmate said that Eve was "so useless"). 

Disclaimer: The images are from an old MMORPG named <a href="https://elsword.koggames.com/">Elsword</a>. I take no credit.

## Getting Started
To get started, you'll need to <a href="https://discord.com/developers/docs/intro">sign up</a> to become a Discord developer, create a bot (application), then get your token. You'll also need a <a href="https://www.mongodb.com/cloud">MongoDB</a> Database Cluster URI, which you can find under SECURITY -> Database Access -> Connect -> Connect your application.

### Deployment
Once you have your token, if you are deploying locally, create a .env file with the following:
```
MONGO_URI = "mongodb+srv://<username>:<password>@<clustername>.<something>.mongodb.net/?retryWrites=true&w=majority"
TOKEN = "yourtokenhere"
```

Then, you can run:
```
go run mary.go
```

### Prerequisites
```
github.com/bwmarrin/discordgo v0.26.1
github.com/joho/godotenv v1.4.0
go.mongodb.org/mongo-driver v1.11.0
```

## Demo
<img src="https://github.com/Chubbyman2/mary-bot/blob/main/docs/demo-1.PNG">

## Built With
### DiscordGo
<a href="https://github.com/bwmarrin/discordgo">DiscordGo</a> is a Go package that provides low level bindings to the Discord chat client API. DiscordGo has nearly complete support for all of the Discord API endpoints, websocket interface, and voice interface. The backbone for this entire project.

### MongoDB Go Driver
The <a href="https://github.com/mongodb/mongo-go-driver">MongoDB Go Driver</a> allows me to store and retrieve data from a MongoDB noSQL database using Go. This is the backbone for the economy system.

## License
This project is licensed under the MIT License - see the <a href="https://github.com/Chubbyman2/mary-bot/blob/main/LICENSE">LICENSE</a> file for details.
