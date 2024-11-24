package main

import (
    //"fmt"
    "github.com/bwmarrin/discordgo"
    //"encoding/json"
    "log"
    //"time"
    "OreSec-bot/util"
    "os"
    "os/signal"
    "syscall"
)

var ticketMessage string 

func main() {
    config, err := util.LoadConfig(".")
    if err != nil {
        log.Fatal("Cannot load config:", err)
    }

    oresecBot, err := discordgo.New(config.DiscordToken)
    if err != nil {
        log.Fatal("Error connecting to bot api", err)
    }

    oresecBot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Print("Bot is up!")
        ticketMessage = util.SendSupportMessage(s, config.TicketChannel)
    })

    oresecBot.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate){
        util.DMMessageCreate(s, m, config.GuildID)
    })

    oresecBot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
            case discordgo.InteractionApplicationCommand:
                if h, ok := util.CommandsHandlers[i.ApplicationCommandData().Name]; ok {
                    h(s, i, config)
                }
            case discordgo.InteractionModalSubmit:
                if h, ok := util.ResponseHandlers[i.ModalSubmitData().CustomID]; ok{
                    h(s, i, config)
                }
            case discordgo.InteractionMessageComponent:
                if h, ok := util.MessageComponentHandler[i.MessageComponentData().CustomID]; ok{
                    h(s, i, config)
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

    go StartPolling(config.CTFDAddress, config.CTFDToken, config.FirstBloodFile, oresecBot, config.FirstBloodChannel, config.BloodRole)
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill) 
    <-stop

    log.Print(ticketMessage)
	if ticketMessage != "" && config.TicketChannel != "" {
		err := oresecBot.ChannelMessageDelete(config.TicketChannel, ticketMessage)
		if err != nil {
			log.Print("Error deleting message:", err)
		} else {
			log.Print("Cleanup successful: support message deleted")
		}
	}
    
    log.Print("Graceful shutdown")

}

