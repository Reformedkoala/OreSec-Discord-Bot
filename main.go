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

    oresecBot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Println("Bot is up!")
    })

    oresecBot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate){
        util.DMMessageCreate(s, m, config.GuildID)
    })

    oresecBot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := util.CommandsHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i)
			}
		case discordgo.InteractionModalSubmit:
            if h, ok := util.ResponseHandlers[i.ModalSubmitData().CustomID]; ok{
                h(s, i, config.CTFDToken, config.CTFDAddress)
            }
		}
	})

    
	cmdIDs := make(map[string]string, len(util.Commands))

	for _, cmd := range util.Commands {
		rcmd, err := oresecBot.ApplicationCommandCreate(config.AppID, config.GuildID, &cmd)
		if err != nil {
			log.Fatalf("Cannot create slash command %q: %v", cmd.Name, err)
		}

		cmdIDs[rcmd.ID] = rcmd.Name
	}

    err = oresecBot.Open() 
    if err != nil {
        log.Fatal("Error opening connection,", err)
        return
    }
    defer oresecBot.Close()

    //go StartPolling(config.CTFDAddress, config.CTFDToken, config.FirstBloodFile, oresecBot)
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill) 
    <-stop

    log.Println("Graceful shutdown")

}

