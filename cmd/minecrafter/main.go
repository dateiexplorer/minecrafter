package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/dateiexplorer/minecrafter/internal/server"
)

// InteractionResponseDataFlagEphemeral is a flag that is officially supported
// by Discord but isn't in the discordgo library yet.
// Only caller of the command can see message.
const InteractionResponseDataFlagEphemeral uint64 = 1 << 6

// conf stores all global variables to configure this bot.
var conf *config

// init function is called before the main function. Loads the configuration
// from an '.env' file. If any error occurred, e.g. because of missing
// variables the error is logged and the applications terminates.
func init() {
	c, err := load()
	if err != nil {
		log.Fatalf("failed loading configuration: %v\n", err)
	}

	// Initialize global variable with conf.
	conf = c
}

type member discordgo.Member

func (m *member) hasRoles(s *discordgo.Session, e *discordgo.InteractionCreate, roles ...string) (bool, error) {
	mRoles := make([]string, 0, len(m.Roles))
	for _, id := range m.Roles {
		role, err := s.State.Role(e.GuildID, id)
		if err != nil {
			return false, fmt.Errorf("failed to get roles: %w", err)
		}

		mRoles = append(mRoles, strings.ToLower(role.Name))
	}

	for _, neededRole := range roles {
		hasRole := false
		for _, role := range mRoles {
			if neededRole == role {
				hasRole = true
			}
		}
		if !hasRole {
			return false, nil
		}
	}

	return true, nil
}

func main() {
	// Initialize the discord session. Add "Bot " before token as recommended
	// in the docs.
	s, err := discordgo.New("Bot " + conf.Token)
	if err != nil {
		log.Fatalf("invalid bot parameters: %v", err)
	}

	// Execute on Ready event
	s.AddHandler(func(s *discordgo.Session, e *discordgo.Ready) {
		log.Println("bot is ready: ", e.User.ID)

		// After bot is ready, check if current server is already up.
		// Set Status with the servers name.
		//
		// Ignore any errors that occurr and don't use the mcping result.
		// Just get the status.
		ser, _ := server.FromServerName(conf.PaperHomePath, "current", conf.ServerIP)
		if status, _, _ := ser.Status(); status == server.Starting || status == server.Up {
			s.UpdateGameStatus(0, ser.Name())
		}
	})

	// Open session connection
	err = s.Open()
	if err != nil {
		log.Fatalf("failed to open session: %v", err)
	}
	defer s.Close()

	command := &discordgo.ApplicationCommand{
		Name:        "miner",
		Description: "Manage minecraft servers",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "info",
				Description: "Get information about the server",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
			{
				Name:        "craft",
				Description: "Start the server",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
			},
		},
	}

	// Register new commands
	_, err = s.ApplicationCommandCreate(s.State.User.ID, conf.GuildID, command)
	if err != nil {
		log.Fatalf("cannot create '%v' command: %v", command.Name, err)
	}

	// Execute if an interaction was created by a user or member
	s.AddHandler(func(s *discordgo.Session, e *discordgo.InteractionCreate) {
		// Block direct messages for user.
		// This should never happened, because Slash commands are not available
		// in private chat but maybe this feature will be introduced later.
		if e.User != nil {
			log.Printf("user '%v' with id '%v' tried to interact with the bot\n", e.User.Username, e.User.ID)

			s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Sorry, but I don't respond to private messages. :thinking:",
					Flags:   InteractionResponseDataFlagEphemeral,
				},
			})
			return
		}

		// Check user permissions
		// Need a valid member for the interaction with a valid role to avoid
		// missuse of this bot.
		mem := e.Member
		if mem == nil {
			log.Println("no member available")

			s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Huhh? You not the type of user I've expected. :thinking:",
					Flags:   InteractionResponseDataFlagEphemeral,
				},
			})
			return
		}

		log.Printf("member '%v' with id '%v' tries to execute command '%v'\n", mem.User.Username, mem.User.ID, command.Options[0].Name)
		var hasPermissions bool
		{
			mem := member(*mem)
			hasPermissions, err = mem.hasRoles(s, e, "miner")
			if err != nil {
				s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Arrg, something went wrong. Please contact the server admin.",
						Flags:   InteractionResponseDataFlagEphemeral,
					},
				})
				return
			}
		}

		if !hasPermissions {
			log.Printf("member '%v' with id '%v' has no permissions\n", mem.User.Username, mem.User.ID)
			s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Oops, you don't have permissions to do that. :thinking:",
					Flags:   InteractionResponseDataFlagEphemeral,
				},
			})
			return
		}

		// At this point the member is valid and has the permissions to execute
		// the commands of the bot. Create a response to avoid interaction failure.
		s.InteractionRespond(e.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: InteractionResponseDataFlagEphemeral,
			},
		})

		if e.ApplicationCommandData().Name == command.Name {
			switch e.ApplicationCommandData().Options[0].Name {
			case "info":
				info(s, e)
			case "craft":
				craft(s, e)
			default:
				s.FollowupMessageCreate(s.State.User.ID, e.Interaction, true, &discordgo.WebhookParams{
					Content: "Oops, I don't know this command yet :thinking:",
					Flags:   InteractionResponseDataFlagEphemeral,
				})
			}
		}
	})

	// Start watcher as a background job
	watcher(s, conf.WatchInterval)

	// Shutdown on interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
