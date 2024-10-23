package main

import (
    "log"
    "strings"
    "github.com/bwmarrin/discordgo"
)

func CommandHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore messages from the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }

    // Simple ping command
    log.Print(m.Content)
    if strings.Contains(m.Content, "!ping") {
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    }
    
    // Add more commands as needed
}

