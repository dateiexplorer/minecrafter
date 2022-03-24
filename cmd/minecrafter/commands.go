package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dateiexplorer/minecrafter/internal/server"
)

func info(s *discordgo.Session, e *discordgo.InteractionCreate) *discordgo.Message {
	ser, err := server.FromServerName(conf.PaperHomePath, "current", conf.ServerIP)
	if err != nil {
		m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
			Content: "Huhh? Something went wrong. Please inform the administrator. :cold_sweat:",
			Flags:   InteractionResponseDataFlagEphemeral,
		})
		return m
	}
	status, res, err := ser.Status()
	if err != nil {
		m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
			Content: "Huhh? Something went wrong. Please inform the administrator. :cold_sweat:",
			Flags:   InteractionResponseDataFlagEphemeral,
		})
		return m
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   "IP",
			Value:  fmt.Sprintf("`%v`", conf.ServerIP),
			Inline: true,
		},
		{
			Name:   "Status",
			Value:  string(status),
			Inline: true,
		},
	}

	if res != nil {
		players := new(strings.Builder)
		for _, player := range res.Sample {
			players.WriteString(fmt.Sprintf("`%v`\n", player.Name))
		}

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   "Players capacity",
			Value:  fmt.Sprintf("%v/%v", res.Online, res.Max),
			Inline: false,
		})

		if players.Len() > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:   "Players online",
				Value:  players.String(),
				Inline: false,
			})
		}
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x6bca5c,
		Title:       fmt.Sprintf("Server information: %v", ser.Name()),
		Description: "Hey, here are some information about the server.",
		Fields:      fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://icons.iconarchive.com/icons/chrisl21/minecraft/256/3D-Grass-icon.png",
		},
	}

	m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  InteractionResponseDataFlagEphemeral,
	})
	return m
}

func craft(s *discordgo.Session, e *discordgo.InteractionCreate) *discordgo.Message {
	ser, err := server.FromServerName(conf.PaperHomePath, "current", conf.ServerIP)
	if err != nil {
		log.Printf("failed to get server by name '%v': %v\n", "current", err)
		m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
			Content: "Huhh? Something went wrong. Please inform the administrator. :cold_sweat:",
			Flags:   InteractionResponseDataFlagEphemeral,
		})
		return m
	}
	status, _, err := ser.Status()
	if err != nil {
		log.Printf("failed to get server status: %v\n", err)
		m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
			Content: "Huhh? Something went wrong. Please inform the administrator. :cold_sweat:",
			Flags:   InteractionResponseDataFlagEphemeral,
		})
		return m
	}

	var title string
	var description string

	switch status {
	case server.Down:
		// Server can be started
		err := ser.Run(conf.CLIExecutable)
		log.Printf("run server '%v'\n", ser.Name())
		if err != nil {
			log.Printf("failed to run server '%v': %v\n", ser.Name(), err)
			m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
				Content: "Huhh? Something went wrong. Please inform the administrator. :cold_sweat:",
				Flags:   InteractionResponseDataFlagEphemeral,
			})
			return m
		}

		title = fmt.Sprintf("Start %v", ser.Name())
		description = "Start the server, it will available shortly... :man_running:"
	case server.Up:
		title = fmt.Sprintf("Server information: %v", ser.Name())
		description = "The server is already up. I'll go sleeping... :sleeping:"
	case server.Starting:
		title = fmt.Sprintf("Server information: %v", ser.Name())
		description = "I started the server, but it's not ready yet. Take a cup of tea... :tea:"
	case server.Stopping:
		title = fmt.Sprintf("Server information: %v", ser.Name())
		description = "Server is shutting down currently. Wait until finished and retry it later. :timer:"
	case server.Locked:
		m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
			Embeds: []*discordgo.MessageEmbed{{
				Author:      &discordgo.MessageEmbedAuthor{},
				Color:       0xad98b1,
				Title:       fmt.Sprintf("Maintenance: %v", ser.Name()),
				Description: "Server is currently locked for maintenance. :pensive: During this I can't start the server. Ask me later.",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://icons.iconarchive.com/icons/chrisl21/minecraft/256/3D-Mycelium-icon.png",
				},
			}},
			Flags: InteractionResponseDataFlagEphemeral,
		})
		return m
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0xad98b1,
		Title:       title,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "IP",
				Value:  fmt.Sprintf("`%v`", ser.IP()),
				Inline: true,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://icons.iconarchive.com/icons/chrisl21/minecraft/256/3D-Mycelium-icon.png",
		},
	}

	m, _ := s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  InteractionResponseDataFlagEphemeral,
	})
	return m
}
