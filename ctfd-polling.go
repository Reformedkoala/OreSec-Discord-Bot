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

func StartPolling(url string, token string, mapfile string, oresecBot *discordgo.Session) {
    file, err := os.Open(mapfile)
    if err != nil {
        log.Fatal(err.Error())
    }

    defer file.Close()
	decoder := gob.NewDecoder(file)

    solveSet := make(map[int]bool)

	err = decoder.Decode(&solveSet)
	if err != nil {
		log.Fatal("failed to decode file ", err.Error())
	}

    ticker := time.NewTicker(5 * time.Second) 
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            pollAPI(url, token, oresecBot, solveSet)
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

func pollAPI(url string, token string, oresecBot *discordgo.Session, solveSet map[int]bool) {

    log.Print("Updating")

    req, err := http.NewRequest("GET", url + "challenges", nil)
    if err != nil {
        log.Fatal("Failed to create GET request ", err.Error())
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", token)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatal("Failed to create client request ", err.Error())
    }

    defer res.Body.Close()
    body, readErr := io.ReadAll(res.Body)
    if readErr != nil {
        log.Fatal(err.Error())
    }

    var response util.Challenges
    json.Unmarshal(body,&response)

    for _, challenge := range response.Data {
        if challenge.Solves != 0{
            _, exists := solveSet[challenge.Id] 
            if !exists {
                req, err = http.NewRequest("GET", url + "challenges/" + fmt.Sprint(challenge.Id) + "/solves", nil)
                req.Header.Set("Content-Type", "application/json")
                req.Header.Set("Authorization", token)
                res, err = http.DefaultClient.Do(req)
                if err != nil {
                    log.Print(err.Error())
                }

                defer res.Body.Close()
                
                body, readErr = io.ReadAll(res.Body)
                if readErr != nil {
                    log.Print(err.Error())
                }
                var solveReturned util.Solves
                json.Unmarshal(body, &solveReturned)
                oresecBot.ChannelMessageSend("982156198243614723", "@Competitors " + solveReturned.Data[0].TeamName + " got a First Blood on \"" + challenge.Name + "\" in the category, " + challenge.Category +"!")                

                solveSet[challenge.Id] = true
            }
        }
    }
    
}

