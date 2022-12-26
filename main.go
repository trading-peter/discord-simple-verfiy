package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/kataras/golog"
	"gopkg.in/yaml.v3"
)

type cfgType struct {
	Token         string `yaml:"token"`
	VerifyRoleID  int    `yaml:"verify_role_id"`
	VerifyChannel string `yaml:"verify_channel"`
	Message       string `yaml:"message"`
}

var cfg cfgType = cfgType{}

func main() {
	yfile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		golog.Fatal(err)
	}

	err = yaml.Unmarshal(yfile, &cfg)
	if err != nil {
		golog.Fatal(err)
	}

	d, err := discordgo.New("Bot " + cfg.Token)

	if err != nil {
		golog.Fatal(err)
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "mass-assign",
			Description: "Assigns the verified role to all existing users",
		},
		{
			Name:        "setup-msg",
			Description: "Send the verification message that visitors have to react to",
		},
		{
			Name:        "update-msg",
			Description: "Update the verification message that visitors have to react to",
		},
	}

	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"mass-assign": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			roleStr := fmt.Sprintf("%d", cfg.VerifyRoleID)
			lastId := ""
			c := 0
			f := 0

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Assigning verified role to existing users",
				},
			})

			for {
				members, err := s.GuildMembers(i.GuildID, lastId, 1000)

				if err != nil {
					golog.Errorf("Failed to fetch member list: %v", err)
					break
				}

				for m := range members {
					f++
					if !hasRole(members[m].Roles, roleStr) {
						golog.Infof("Added verified role to %s", members[m].User.Username)
						err := s.GuildMemberRoleAdd(i.GuildID, members[m].User.ID, roleStr)

						if err != nil {
							golog.Errorf("Failed to assign role to member %s: %v", members[m].User.Username, err)
							break
						}

						c++
					}
				}

				if len(members) < 1000 {
					break
				}

				lastId = members[len(members)-1].User.ID
			}

			golog.Infof("Assigned role to %d/%d users", c, f)
		},
		"setup-msg": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			_, err := s.ChannelMessageSend(cfg.VerifyChannel, cfg.Message)

			if err != nil {
				golog.Error(err)
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "☑",
				},
			})
		},
		"update-msg": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			list, err := s.ChannelMessages(cfg.VerifyChannel, 100, "", "", "")

			if err != nil {
				golog.Error(err)
			}

			for i := range list {
				if list[i].Author.ID == s.State.User.ID {
					s.ChannelMessageEdit(cfg.VerifyChannel, list[i].ID, cfg.Message)
					break
				}
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "☑",
				},
			})
		},
	}

	d.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	d.AddHandler(messageCreate)

	err = d.Open()
	if err != nil {
		golog.Errorf("Error opening connection: %v", err)
		return
	}

	for _, v := range commands {
		_, err := d.ApplicationCommandCreate(d.State.User.ID, "", v)
		if err != nil {
			golog.Fatalf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

	fmt.Println("Press CTRL+C to stop the bot")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	d.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.Emoji.Name == "👍" && m.ChannelID == cfg.VerifyChannel && !hasRole(m.Member.Roles, fmt.Sprintf("%d", cfg.VerifyRoleID)) {
		err := s.GuildMemberRoleAdd(m.GuildID, m.Member.User.ID, fmt.Sprintf("%d", cfg.VerifyRoleID))
		if err != nil {
			golog.Errorf("Failed to assigned verified role to user %s: %v", m.Member.User.ID, err)
		}
	}
}

func hasRole(list []string, role string) bool {
	for v := range list {
		if list[v] == role {
			return true
		}
	}
	return false
}
