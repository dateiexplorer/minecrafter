package main

import (
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dateiexplorer/minecrafter/internal/server"
)

func watcher(s *discordgo.Session, interval time.Duration) {
	go func() {
		counter := 0
		var last *server.Server
		for {
			<-time.Tick(interval)

			// Get the current server and make a mcping command to its ip.
			ser, err := server.FromServerName(conf.PaperHomePath, "current", conf.ServerIP)
			if err != nil {
				log.Printf("failed to get server by name '%v': %v\n", "current", err)
				continue
			}

			// If the server of the last interval is'nt the same, stop the last server.
			if last != nil && *ser != *last {
				log.Printf("found previous running server '%v': shutting down\n", last.Name())
				last.Stop(conf.CLIExecutable)
			}
			last = ser

			// Get the status of the current server.
			status, res, err := ser.Status()
			if err != nil {
				counter++
				log.Printf("watcher: error: increase counter to %v/%v\n", counter, conf.MaxAttempts)
			} else {
				switch status {
				case server.Down, server.Locked, server.Stopping, server.Starting:
					counter = 0
				case server.Undefined:
					counter++
					log.Printf("watcher: status UNDEFINDED: increase counter to %v/%v\n", counter, conf.MaxAttempts)
				case server.Up:
					s.UpdateGameStatus(0, ser.Name())
					// If no players on the server and server is'nt currently starting
					// increase the counter.
					if res == nil || res.Online == 0 {
						counter++
						log.Printf("watcher: no players on the server: increase counter to %v/%v\n", counter, conf.MaxAttempts)
					} else {
						counter = 0
						log.Printf("watcher: %v players online: reset counter to %v/%v\n", res.Online, counter, conf.MaxAttempts)
					}
				}
			}

			if counter >= conf.MaxAttempts {
				log.Printf("watcher: max attempts reached, stop server %v\n", ser.Name())
				ser.Stop(conf.CLIExecutable)
				s.UpdateGameStatus(0, "")
				counter = conf.MaxAttempts
			}
		}
	}()
}
