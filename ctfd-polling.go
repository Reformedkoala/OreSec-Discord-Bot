package main 

import (
    "fmt"
    "time"
)

func StartPolling(url string) {
    ticker := time.NewTicker(30 * time.Second) // Poll every 30 seconds
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            pollAPI(url)
        }
    }
}

func pollAPI(url string) {

    fmt.Println("Live Update")

    // Send the updates to a Discord channel if needed
    // You can fetch channel ID dynamically or hardcode it
}

