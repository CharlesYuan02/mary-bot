# mary-bot
<img src="https://i.ytimg.com/vi/9S831972rjA/maxresdefault.jpg">

## Introduction
A Discord bot I created in Go. Her name is Mary. After receiving my offer to work at Uber, I knew I had to learn Go right away, as my interviewer informed me that that was the primary language they used for back end development. I decided to learn some noSQL and MongoDB to implement a database-driven game so that people will actually interact with her (cuz my classmate said that Eve was "so useless"). 

Want to invite her to your server? Use <a href="https://discord.com/api/oauth2/authorize?client_id=1038557818200019025&permissions=8&scope=bot">this link</a>.

Disclaimer: The images are from an old MMORPG named <a href="https://elsword.koggames.com/">Elsword</a>. I take no credit.

## Getting Started
To get started, you'll need to <a href="https://discord.com/developers/docs/intro">sign up</a> to become a Discord developer, create a bot (application), then get your token. You'll also need a <a href="https://www.mongodb.com/cloud">MongoDB</a> Database Cluster URI, which you can find under SECURITY -> Database Access -> Connect -> Connect your application. Remember to whitelist your IP Address or allow all IP addresses if you're hosting!

### Prerequisites
```
github.com/bwmarrin/discordgo v0.26.1
github.com/joho/godotenv v1.4.0
go.mongodb.org/mongo-driver v1.11.0
github.com/asaskevich/govalidator v0.0.0-20210307081110-f21760c49a8d
```

### Local Deployment
Once you have your token, if you are deploying locally, create a .env file with the following:
```
MONGO_URI = "mongodb+srv://<username>:<password>@<clustername>.<something>.mongodb.net/?retryWrites=true&w=majority"
OWNER_ID = "yourDiscordUserID"
TOKEN = "yourtoken"
```

Then, you can run:
```
go run mary.go
```

### Deployment on Google Cloud Virtual Machine
First, if you haven't already, you'll need to create a <a href="https://cloud.google.com/">Google Cloud</a> account and enable the Compute Engine API. Follow the first part of <a href="https://cloud.google.com/blog/topics/developers-practitioners/build-and-run-discord-bot-top-google-cloud">these instructions</a> if you need help. After that, you will need to <a href="https://medium.com/@emerson15dias/how-to-install-go-on-a-vm-virtual-box-running-ubuntu-under-windows-988ce34329eb">set up dependencies</a> on your virtual machine (i.e. wget, git, Go, tmux):
```
$ sudo apt-get install wget
$ sudo apt-get install tmux
$ wget https://storage.googleapis.com/golang/go1.19.linux-amd64.tar.gz
$ sudo tar -xvf go1.19.linux-amd64.tar.gz
$ sudo mv go /usr/local
$ sudo apt install git
```
Then, clone the repo and find your GOPATH directory:
```
$ git clone https://github.com/Chubbyman2/mary-bot.git
$ pwd
```
Next, create a .env file with your environment variables:
```
$ cd mary-bot
$ sudo nano .env 
### Copy-paste your env vars into this file, CTRL + X, Enter to save
```
Afterwards, set up your go environment in a tmux session (see <a href="https://www.youtube.com/watch?v=VEn70C7S5Q8">this tutorial</a> for details).
```
$ tmux
$ export GOROOT=/usr/local/go
### Set GOPATH=[response from pwd], mine is /home/charlesyuan59
$ export GOPATH=/home/charlesyuan59
$ export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
```
Finally, you can run Mary in the tmux session:
```
$ go run mary.go
### CTRL+B (hold), then D to exit out of tmux session
```
Additionally, here are some useful commands for tmux:
```
### Show tmux instances and their id numbers
$ tmux ls 
### Join tmux instance
$ tmux a -t [id]
### Kill tmux session
$ tmux kill-session
```

## Built With
### DiscordGo
<a href="https://github.com/bwmarrin/discordgo">DiscordGo</a> is a Go package that provides low level bindings to the Discord chat client API. DiscordGo has nearly complete support for all of the Discord API endpoints, websocket interface, and voice interface. The backbone for this entire project.

### MongoDB Go Driver
The <a href="https://github.com/mongodb/mongo-go-driver">MongoDB Go Driver</a> allows me to store and retrieve data from a MongoDB noSQL database using Go. This is the backbone for the economy system.

### Google Cloud + tmux
<a href="https://cloud.google.com/">Google Cloud</a>'s Compute Engine provides me with a virtual machine instance that I can ssh into, download my dependencies, and host Mary from. I then used <a href="https://en.wikipedia.org/wiki/Tmux">tmux</a> to run the code in the background, so that when I close the secure shell window, Mary keeps running.

## Update March 4, 2023
My GCP trial ran out, so I've switched to using <a href="https://fly.io/">Fly.io/</a>. I've also been working for the past two weeks on more economy functions. This includes trivia, inventory, shop, give (an item, not coins), buy, sell, and five different use commands. The help, trivia, inventory, shop, and profile commands are all sent as rich embeds now, too! Feel free to try them out.

## Demo
<img src="https://github.com/Chubbyman2/mary-bot/blob/main/docs/demo-2.PNG">

## Future Plans
### StockInfo
Basically, I want to add a basic stock trading feature to Mary. No stops or limits, just buy and sell using the Yahoo Finance API. The implementation will be similar to what I did with trivia (API call) and inventory (portfolio of stocks). StockInfo will retrieve the current value of the stock along with other relevant information. 

### BuyStock, SellStock
Buy and sell a stock at the market price.

### Portfolio
Portfolio will display the user's current stock porfolio, the value of each stock currently, the original purchase price, and the overall performance.

### MaryPortfolio
A school project I am doing is a sentiment analysis stock trader, which I plan on turning into an API. I will let Mary have her own trading portfolio using the API for buy and sell suggestions.

## License
This project is licensed under the MIT License - see the <a href="https://github.com/Chubbyman2/mary-bot/blob/main/LICENSE">LICENSE</a> file for details.
