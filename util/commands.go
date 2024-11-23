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
                    Required:    true, // Mark this argument as required
                },
            },
		},
	}

var CommandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, token string, url string){
		"create_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, token string, url string) {
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
        "get_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, token string, url string){
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

var ResponseHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, token string, url string){
		"create_challenge": func(s *discordgo.Session, i *discordgo.InteractionCreate, token string, url string) {    
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
            
            log.Printf("ID: %d Challenge Name: %s, Catgeory: %s, Description: %s, Initial: %d, Decay: %d, Final: %d, Flag: %s", 
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
