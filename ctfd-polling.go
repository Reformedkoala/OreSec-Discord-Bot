package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"OreSec-bot/util"
	"github.com/bwmarrin/discordgo"
    "encoding/gob"
    "os"
)

func StartPolling(url string, token string, mapfile string, oresecBot *discordgo.Session, firstBloodChannel string, bloodRole string) {
    file, err := os.Open(mapfile)
    if err != nil {
        file, err = os.Create(mapfile)
        if err != nil {
            fmt.Println("Creating new Gob file", err)
            return
        }
    }

    defer file.Close()
	decoder := gob.NewDecoder(file)

    solveSet := make(map[int]bool)

	err = decoder.Decode(&solveSet)
	if err != nil {
		log.Print("File empty, writing all new challenges ", err.Error())
	}

    ticker := time.NewTicker(10 * time.Second) 
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            pollAPI(url, token, oresecBot, solveSet, firstBloodChannel, bloodRole)
            file, err := os.Create(mapfile)
            if err != nil {
                log.Fatal("failed to open file ", err.Error())
            }
            defer file.Close()
            encoder := gob.NewEncoder(file)

            // Encode the map into the file
            err = encoder.Encode(solveSet)
            if err != nil {
                log.Fatal("Failed to encode results ", err.Error())
            }
        }
    }
}

func pollAPI(url string, token string, oresecBot *discordgo.Session, solveSet map[int]bool, firstBloodChannel string, bloodRole string) {
    req, err := http.NewRequest("GET", url + "challenges", nil)
    if err != nil {
        log.Print("Failed to create GET request ", err.Error())
        return
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", token)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Print("Failed to create client request ", err.Error())
        return
    }

    defer res.Body.Close()
    body, readErr := io.ReadAll(res.Body)
    if readErr != nil {
        log.Print(err.Error())
        return
    }

    var response util.Challenges
    json.Unmarshal(body,&response)

    for _, challenge := range response.Data {
        if challenge.Solves != 0 {
            _, exists := solveSet[challenge.Id]
            if !exists {
                req, err = http.NewRequest("GET", url + "challenges/" + fmt.Sprint(challenge.Id) + "/solves", nil)
                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("Authorization", token)
                res, err = http.DefaultClient.Do(req)
                if err != nil {
                    log.Print(err.Error())
                    return
                }

                defer res.Body.Close()
                
                body, err = io.ReadAll(res.Body)
                if err != nil {
                    log.Print(err.Error())
                    return
                }
                var solveReturned util.Solves
                json.Unmarshal(body, &solveReturned)
                message := fmt.Sprintf("<@&%s> %s got a First Blood on \"%s\" in the category, %s!", bloodRole, solveReturned.Data[0].TeamName, challenge.Name, challenge.Category)
                _, err = oresecBot.ChannelMessageSend(firstBloodChannel, message)                
                if err != nil {
                    log.Print(err.Error())
                    return
                }
                log.Print(message)
                solveSet[challenge.Id] = true
            }
        }
    }
    
}

