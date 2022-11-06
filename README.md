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
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/joho/godotenv v1.4.0 // direct
	golang.org/x/crypto v0.0.0-20210421170649-83a5a9bb288b // indirect
	golang.org/x/sys v0.0.0-20201119102817-f84b799fce68 // indirect
)
```

## License
This project is licensed under the MIT License - see the <a href="https://github.com/Chubbyman2/mary-bot/blob/main/LICENSE">LICENSE</a> file for details.
