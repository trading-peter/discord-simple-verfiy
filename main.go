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

	d.AddHandler(messageCreate)

	err = d.Open()
	if err != nil {
		golog.Errorf("Error opening connection: %v", err)
		return
	}

	fmt.Println("Press CTRL+C to stop the bot")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	d.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.Emoji.Name == "👍" && m.ChannelID == cfg.VerifyChannel && !hasVerifiedRole(m.Member.Roles) {
		err := s.GuildMemberRoleAdd(m.GuildID, m.Member.User.ID, fmt.Sprintf("%d", cfg.VerifyRoleID))
		if err != nil {
			golog.Errorf("Failed to assigned verified role to user %s: %v", m.Member.User.ID, err)
		}
	}
}

func hasVerifiedRole(list []string) bool {
	for v := range list {
		if v == cfg.VerifyRoleID {
			return true
		}
	}
	return false
}
