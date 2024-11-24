package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"github.com/bwmarrin/discordgo"
)

var TicketCount = 0

var TicketTracking = make(map[string]*TicketData)

var Commands = []discordgo.ApplicationCommand{
		{
			Name:        "create_challenge",
			Description: "Create a CTF Challenge",
		},
        {
            Name:        "get_challenge",
            Description: "Provide a challenge id and return the challenge information",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionInteger,
                    Name:        "id",
                    Description: "A required id",
                    Required:    true, 
                },
            },
		},
	}

var CommandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
		"create_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "create_challenge",
					Title:    "Challenge Creation",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "Name",
                                    Label:       "Name",
									Style:       discordgo.TextInputShort,
									Placeholder: "Make sure it's creative!",
									Required:    true,
									MaxLength:   64,
									MinLength:   1,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "Category",
                                    Label:     "Category",
									Style:     discordgo.TextInputShort,
                                    Placeholder: "If you need help determining category, please ask!",
									Required:  true,
									MaxLength: 64,
                                    MinLength: 1,
								},
							},
						},
                        discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "Message",
									Label:     "Message",
									Style:     discordgo.TextInputParagraph,
                                    Placeholder: "Come up with a good description and introduction to your challenge!",
									Required:  true,
									MaxLength: 2000,
                                    MinLength: 10,
								},
							},
						},
                        discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "Points",
									Label:     "Initial Value, Decay, Minimum Value",
									Style:     discordgo.TextInputShort,
                                    Placeholder: "100,10,50 Initial value of 100, decaying until 10 teams, minimum value of 50.",
									Required:  true,
								},
							},
						},
                        discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:  "Flag",
									Label:     "Flag",
									Style:     discordgo.TextInputShort,
                                    Placeholder: "This is the flag for your challenge",
									Required:  true,
                                    MinLength: 1,
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
        "get_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
            url := config.CTFDAddress
            token := config.CTFDToken
            id := i.ApplicationCommandData().Options[0].IntValue()
            response, err := WebsiteRequest(url, token, "challenges", "GET", "/" + fmt.Sprintf("%d", id), "")
            if err != nil {
                err = Respond(s, i, "Unable to fetch challenge, please try again or message an Admin")

                if err != nil {
                    panic(err)    
                }
            }
            challenge := response.(*GenericResponse)
            if !challenge.Success {
                err = Respond(s, i, fmt.Sprintf("Unable to fetch challenge ID %d, please try a different ID or message an Admin", id)) 

                if err != nil {
                    panic(err)    
                }
            }
            response, err = WebsiteRequest(url, token, "flags", "GET", "/" + fmt.Sprintf("%d", id), "")
            if err != nil {
                err = Respond(s, i, "Unable to fetch flag, please try again or message an Admin")

                if err != nil {
                    panic(err)    
                }
            }
            flag := response.(*GenericResponse)
            err = Respond(s, i, fmt.Sprintf("Challenge ID: %.0f\nChallenge Name: %s\nChallenge Category %s\nChallenge Description: %s \nInitial Value: %.0f\nDecay: %.0f\nMinimum: %.0f\nFlag: %s", 
                challenge.Data["id"], 
                challenge.Data["name"], 
                challenge.Data["category"], 
                challenge.Data["description"], 
                challenge.Data["initial"], 
                challenge.Data["decay"], 
                challenge.Data["minimum"],
                flag.Data["content"])) 
            if err != nil {
                panic(err)    
            }
        },
    }

var ResponseHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
		"create_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config) {    
            url := config.CTFDAddress
            token := config.CTFDToken
			data := i.ModalSubmitData()

			if !strings.HasPrefix(data.CustomID, "create_challenge") {
				return
			}
            var challenge ChallengePost 
            challenge.Name = data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
            challenge.Category = data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value 
            challenge.Description = data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value 
			points := data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value 
            flag := data.Components[4].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value 
		    
            trimmed := strings.ReplaceAll(points, " ", "")
	        parts := strings.Split(trimmed, ",")
            
            var numbers []int
            for _, part := range parts {
                num, err := strconv.Atoi(part)
                if err != nil {
                    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                        Type: discordgo.InteractionResponseChannelMessageWithSource,
                        Data: &discordgo.InteractionResponseData{
                            Content: "Your challenge was not submitted, please ensure your points input are valid numbers and in the right format.",
                            Flags: discordgo.MessageFlagsEphemeral,
                        },
                    })

                    if err != nil {
                        panic(err)    
                    }
                    return
                }
                numbers = append(numbers, num)
            }
            
            challenge.Initial = numbers[0]
            challenge.Decay = numbers[1]
            challenge.Minimum = numbers[2]
		    challenge.Function = "logarithmic"
            challenge.State = "hidden"
            challenge.Type = "dynamic"
            
            webResponse, err := WebsiteRequest(url, token, "challenges", "POST", "", challenge) 
            submitResponse := webResponse.(*ChallengeSubmit)
            if err != nil {
                log.Print(err.Error())
                err = Respond(s, i, "We ran into an issue sending the challenge to the server, DM an Admin")
                if err != nil {
                    panic(err)
                }
                return
            }

            var flagPost FlagPost 
            flagPost.Challenge_ID = submitResponse.Data.Id
            flagPost.Content = flag
            flagPost.Data = "case_insensitive"
            flagPost.Type = "static"
            
            webResponse, err = WebsiteRequest(url, token, "flags", "POST", "", flagPost)
            if err != nil {
                log.Print(err.Error())
                err = Respond(s, i, "We ran into an issue sending the flag to the server, DM an Admin")
                if err != nil {
                    panic(err)
                }
                return
            }
            
            log.Printf("ID: %d Challenge Name: %s, Category: %s, Description: %s, Initial: %d, Decay: %d, Final: %d, Flag: %s", 
                submitResponse.Data.Id,
                challenge.Name, 
                challenge.Category, 
                challenge.Description, 
                challenge.Initial, 
                challenge.Decay, 
                challenge.Minimum, 
                flag,
            )

            err = Respond(s, i, "Your challenge has been submitted! Please take note of this id in order to make any changes: " + fmt.Sprintf("%d", submitResponse.Data.Id))
            if err != nil {
                panic(err)    
            }
        },
        "ticket_channel_creation": func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config) {
            TicketCount += 1
            channelName := fmt.Sprintf("ticket-%d", TicketCount)
            newChannel, err := s.GuildChannelCreateComplex(config.GuildID, discordgo.GuildChannelCreateData{
                Name:     channelName,
                Type:     discordgo.ChannelTypeGuildText,
                ParentID: config.TicketCategory,
                PermissionOverwrites: []*discordgo.PermissionOverwrite{
                    {
                        ID:   config.GuildID,
                        Type: discordgo.PermissionOverwriteTypeRole,
                        Deny: discordgo.PermissionViewChannel,
                    },
                    {
                        ID:   i.Member.User.ID, 
                        Type: discordgo.PermissionOverwriteTypeMember,
                        Allow: discordgo.PermissionSendMessages |
                            discordgo.PermissionViewChannel,
                    },
                    {
                        ID:   config.VolunteerRole, 
                        Type: discordgo.PermissionOverwriteTypeRole,
                        Allow: discordgo.PermissionViewChannel |
                            discordgo.PermissionSendMessages,
                    },
                },
            }) 

            TicketTracking[newChannel.ID] = &TicketData {
                OpenedBy: i.Member.User.ID,
                PlayerName: i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
                TicketSubject: i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
                TicketDescription: i.ModalSubmitData().Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
                ClosedBy: "",
            }

            if err != nil {
                panic(err)
            }
            _, err = s.ChannelMessageSendComplex(newChannel.ID, &discordgo.MessageSend{
                Embed: &discordgo.MessageEmbed{
                    Title: fmt.Sprintf("%s", i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value),
                    Description: fmt.Sprintf("Thank you for creating an OreSec support ticket.\n\nThe <@&%s>'s will assist you as soon as possible. We appreciate your patience and hope you are enjoying things so far!\n\n While you wait please refer to the FAQ and Information channels. You can also add any other relevant information in this channel.", config.VolunteerRole),
                    Color: 0x1F8B4C,
                },
                Components: []discordgo.MessageComponent{
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.Button{
                                Emoji: discordgo.ComponentEmoji{
                                    Name: "\U0000274C",
                                }, 
                                Label: "Close Ticket",
                                Style: discordgo.PrimaryButton,
                                CustomID: "close_ticket",
                            },
                        },
                    },
                },
            })
            if err != nil {
                fmt.Println("Error sending button message:", err)
            }
            _, err = s.ChannelMessageSendComplex(newChannel.ID, &discordgo.MessageSend{
                Embed: &discordgo.MessageEmbed{
                    Description: "Submitted Information",
                    Color: 0x1F8B4C,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name:   "Player Name",
                            Value:  TicketTracking[newChannel.ID].PlayerName,
                            Inline: false,
                        },
                        {
                            Name:   "Ticket Subject",
                            Value:  TicketTracking[newChannel.ID].TicketSubject,
                            Inline: false,
                        },
                        {
                            Name:   "Description",
                            Value:  TicketTracking[newChannel.ID].TicketDescription,
                            Inline: false,
                        },
                    },
                },
            })
            Respond(s, i, fmt.Sprintf("Your ticket channel has been created and an Admin will take a look shortly, please head over to <#%s>", newChannel.ID))
            return
        }, 
        "ticket_channel_deletion": func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config) {
            channel, err := s.UserChannelCreate(TicketTracking[i.ChannelID].OpenedBy)

            if err != nil {
                log.Print("error creating channel:", err)
                s.ChannelMessageSend(
                    config.TicketChannel,
                    "Something went wrong while sending the DM!",
                )
                return
            }

            _, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
                Embed: &discordgo.MessageEmbed{
                    Title: "Ticket Closed",
                    Description: "Your ticket was closed. Please refer to below for a short summary",
                    Color: 0x1F8B4C,
                    Fields: []*discordgo.MessageEmbedField{
                        {
                            Name:   "Opened By",
                            Value:  fmt.Sprintf("<@%s>", TicketTracking[i.ChannelID].OpenedBy),
                            Inline: false,
                        },
                        {
                            Name:   "Closed By",
                            Value:  fmt.Sprintf("<@%s>", TicketTracking[i.ChannelID].ClosedBy),
                            Inline: false,
                        },
                        {
                            Name:   "Player Name",
                            Value:  TicketTracking[i.ChannelID].PlayerName,
                            Inline: false,
                        },
                        {
                            Name:   "Ticket Subject",
                            Value: TicketTracking[i.ChannelID].TicketSubject,
                            Inline: false,
                        },
                        {
                            Name:   "Description",
                            Value: TicketTracking[i.ChannelID].TicketDescription,
                            Inline: false,
                        },
                    },
                },
            })

            if err != nil {
                log.Print("error sending DM message:", err)
                s.ChannelMessageSend(
                    config.TicketChannel,
                    "Failed to send you a DM. Did you disable DM's in your privacy settings?",
                )
            }
            Respond(s, i, "Deleting Channel and closing the ticket")
            s.ChannelDelete(i.ChannelID)
            return
        }, 
    }

var MessageComponentHandler = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
    "create_ticket":func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseModal,
            Data: &discordgo.InteractionResponseData{
                CustomID: "ticket_channel_creation",
                Title:    "Ticket Creation",
                Components: []discordgo.MessageComponent{
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "name",
                                Label:       "Team Name or Username",
                                Style:       discordgo.TextInputShort,
                                Placeholder: "Enter your ctfd team name or username here.",
                                Required:    true,
                                MaxLength:   64,
                                MinLength:   1,
                            },
                        },
                    },
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "subject",
                                Label:       "Subject",
                                Style:       discordgo.TextInputShort,
                                Placeholder: "Summarize your ticket in a subject line.",
                                Required:    true,
                                MaxLength:   64,
                                MinLength:   1,
                            },
                        },
                    },
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "issue",
                                Label:       "Issue Description",
                                Style:       discordgo.TextInputParagraph,
                                Placeholder: "Describe in detail your problem and we'll attempt to solve it.",
                                Required:    true,
                                MaxLength:   4000,
                                MinLength:   1,
                            },
                        },
                    },

                },
            },
        })
        if err != nil {
            panic(err)
        }
        return
    },
    "close_ticket":func(s *discordgo.Session, i *discordgo.InteractionCreate, config Config){
        if t, ok := TicketTracking[i.ChannelID];ok{
            t.ClosedBy = i.Member.User.ID
        }
        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseModal,
            Data: &discordgo.InteractionResponseData{
                CustomID: "ticket_channel_deletion",
                Title:    "Ticket Closing",
                Components: []discordgo.MessageComponent{
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "Reason",
                                Label:       "Reason",
                                Style:       discordgo.TextInputParagraph,
                                Placeholder: "Enter reason for ticket closure here",
                                Required:    true,
                                MaxLength:   4000,
                                MinLength:   1,
                            },
                        },
                    },
                },
            },
        })
        if err != nil {
            panic(err)
        }
        return
    },
}

func DMMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate, guildID string) {
    // Ignore messages from the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }
    //If the guildID is blank, means no guild from config
    if m.GuildID != "" { 
            return
    }
    //Checking for Create Challenge command
    if !strings.Contains(m.Content, "!CreateChallenge"){
        return
    }
    //Grabbing member information of the correct server, server is controlled via config
    member, err := s.GuildMember(guildID, m.Author.ID)
    if err != nil {
		s.ChannelMessageSend(m.ChannelID, "Error fetching your role information.")
		log.Printf("error fetching member data: %v", err)
		return
	}
    //Looping over roles in the server, unfortunately it's just an array
    //Maybe change this to look for a specific role in the future
    var roleCheck = false 
    for _, roleID := range member.Roles {
        role, err := s.State.Role(guildID, roleID)
        if err != nil{
            continue
        }
        
        if role != nil && "TEST" == role.Name{
            roleCheck = true 
        }
    }
    //Opening DM with user
    log.Print(m.Content)
    channel, err := s.UserChannelCreate(m.Author.ID)
    if err != nil {
        log.Print("error creating channel:", err)
        s.ChannelMessageSend(
            m.ChannelID,
            "Something went wrong while sending the DM!",
        )
        return
    }
    //Ending if Role is not correct, we still want to inform the user
    if !roleCheck {
        log.Print("Incorrect role for user")
        s.ChannelMessageSend(
            m.ChannelID,
            "You don't have the correct role, DM the organizer",
        )
        return
    }
    //Sending the DM Message
    _, err = s.ChannelMessageSend(channel.ID, "Pong!")
    if err != nil {
        log.Print("error sending DM message:", err)
        s.ChannelMessageSend(
            m.ChannelID,
            "Failed to send you a DM. "+
                "Did you disable DM in your privacy settings?",
        )
    }
}


func WebsiteRequest(url string, token string, endpoint string, requestType string, option string, data any) (any, error) {
     
    var submitResponse interface{}

    switch endpoint {
        case "challenges":
            switch requestType {
                case "POST":
                    submitResponse = &ChallengeSubmit{}
                case "GET":
                    submitResponse = &GenericResponse{}
                default:
                    return nil, fmt.Errorf("Invalid requestType: %s", requestType)
            }
        case "flags":
            submitResponse = &GenericResponse{}
        default:
            return nil, fmt.Errorf("Invalid endpoint: %s", endpoint)
    }


    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("Error marshalling JSON: %s", err.Error())
    }

    req, err := http.NewRequest(requestType, url + endpoint + option, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("Failed to create post request: %s", err.Error())
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", token)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("Failed to create client request: %s", err.Error())
    }

    defer res.Body.Close()
    body, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, fmt.Errorf("Failed to read body of request: %s", err.Error())
    }
    
    if err := json.Unmarshal(body, submitResponse); err != nil {
        return nil, fmt.Errorf("Error unmarshalling JSON response: %s\n", err.Error())
    }

    return submitResponse, nil
}

func Respond(s *discordgo.Session, i *discordgo.InteractionCreate, content string) error{
    err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: content,
            Flags: discordgo.MessageFlagsEphemeral,
        },
    })

    return err
}

func SendSupportMessage(s *discordgo.Session, channelID string) string {
    message, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title: "Create a Ticket",
			Description: "Welcome to an OreSec ran competition. This bot is here to help manage any inquiries you may have through tickets.\n\nThis allows you to get direct and discrete feedback from the Admins on any problems you may face.",
			Color: 0x1F8B4C,
		},
		Components: []discordgo.MessageComponent{
            discordgo.ActionsRow{
                Components: []discordgo.MessageComponent{
                    discordgo.Button{
                        Emoji: discordgo.ComponentEmoji{
                            Name: "\U0001F4E9",
                        }, 
                        Label: "Create Ticket",
                        Style: discordgo.PrimaryButton,
                        CustomID: "create_ticket",
                    },
                },
            },
        },
	})
	if err != nil {
		fmt.Println("Error sending button message:", err)
	}
    return message.ID
}
