package main

import (
	//"fmt"
	"github.com/bwmarrin/discordgo"
	//"encoding/json"
	//"fmt"
	//"io"
	"log"
	//"net/http"
	//"time"
	"OreSec-bot/util"
	"os"
	"os/signal"
	"syscall"
)

func main() {
    config, err := util.LoadConfig(".")
    if err != nil {
        log.Fatal("cannot load config:", err)
    }

    oresecBot, err := discordgo.New(config.DiscordToken)
    if err != nil {
        log.Fatal("Error connecting to bot api", err)
    }
   
    oresecBot.AddHandler(CommandHandler)
    err = oresecBot.Open() 
    if err != nil {
        log.Fatal("Error opening connection,", err)
        return
    }
    log.Print("Bot is now running. Press CTRL+C to exit.")
    
    go StartPolling(config.CTFDAddress, config.CTFDToken, config.FirstBloodFile, oresecBot)
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill) 
    <-stop

    // Cleanly close down the Discord session
    oresecBot.Close()
}

