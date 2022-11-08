# mary-bot
<img src="https://i.ytimg.com/vi/9S831972rjA/maxresdefault.jpg">

## Introduction
A Discord bot I created in Go. Her name is Mary. After receiving my offer to work at Uber, I knew I had to learn Go right away, as my interviewer informed me that that was the primary language they used for back end development. I decided to learn some noSQL and MongoDB to implement a database-driven game so that people will actually interact with her (cuz my classmate said that Eve was "so useless"). 

Disclaimer: The images are from an old MMORPG named <a href="https://elsword.koggames.com/">Elsword</a>. I take no credit.

## Getting Started
To get started, you'll need to <a href="https://discord.com/developers/docs/intro">sign up</a> to become a Discord developer, create a bot (application), then get your token. 

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
module mary-bot

go 1.19

require (
	github.com/bwmarrin/discordgo v0.26.1 // direct
	github.com/joho/godotenv v1.4.0 // direct
)

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.11.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.7 // indirect
)
```

## License
This project is licensed under the MIT License - see the <a href="https://github.com/Chubbyman2/mary-bot/blob/main/LICENSE">LICENSE</a> file for details.
