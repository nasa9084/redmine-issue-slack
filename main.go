package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"unicode"

	flags "github.com/jessevdk/go-flags"
	"github.com/lestrrat-go/slack"
	"github.com/lestrrat-go/slack/objects"
	"github.com/lestrrat-go/slack/rtm"
	redmine "github.com/mattn/go-redmine"
)

type options struct {
	Slack   slackOptions
	Redmine redmineOptions
}

type slackOptions struct {
	Token string `short:"t" long:"slack-token" env:"SLACK_TOKEN" required:"true" description:"API Token for slack bot"`
}

type redmineOptions struct {
	Endpoint string `short:"r" long:"redmine-endpoint" env:"REDMINE_ENDPOINT" required:"true" description:"Endpoint of your Redmine"`
	APIKey   string `short:"k" long:"redmine-apikey" env:"REDMINE_APIKEY" required:"true" description:"API Key for your Redmine"`
}

var userMap = loadUserMap()

var opts options
var (
	slackRESTClient *slack.Client
	slackRTMClient  *rtm.Client
	redmineClient   *redmine.Client
)

func main() { os.Exit(_main()) }

func _main() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := exec(ctx); err != nil {
		log.Print(err)
		return 1
	}
	return 0
}

func exec(ctx context.Context) error {
	if _, err := flags.Parse(&opts); err != nil {
		return err
	}
	if err := initClients(ctx); err != nil {
		return err
	}

	go slackRTMClient.Run(ctx)
	listenEvent(ctx)
	return nil
}

func loadUserMap() map[string]string {
	f, err := os.Open("./usermapping.json")
	if err != nil {
		return map[string]string{}
	}
	defer f.Close()
	m := map[string]string{}
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return map[string]string{}
	}
	return m
}

func initClients(ctx context.Context) error {
	slackRESTClient = slack.New(opts.Slack.Token)
	if _, err := slackRESTClient.Auth().Test().Do(ctx); err != nil {
		return err
	}
	slackRTMClient = rtm.New(slackRESTClient)
	redmineClient = redmine.NewClient(opts.Redmine.Endpoint, opts.Redmine.APIKey)
	return nil
}

func listenEvent(ctx context.Context) {
	fmt.Println("listen")
	for e := range slackRTMClient.Events() {
		switch typ := e.Type(); typ {
		case rtm.MessageType:
			processMessage(ctx, e.Data().(*rtm.MessageEvent))
		default:
		}
	}
}

func processMessage(ctx context.Context, e *rtm.MessageEvent) {
	if e.User == "" {
		return
	}
	ticketID := extractTicketID(e.Text)
	if ticketID < 0 {
		return
	}
	issue, err := redmineClient.Issue(ticketID)
	if err != nil {
		return
	}
	msg := fmt.Sprintf("<%s/issues/%d|#%d>: %s", opts.Redmine.Endpoint, ticketID, ticketID, issue.Subject)
	attachment := &objects.Attachment{
		Fields: objects.AttachmentFieldList{
			&objects.AttachmentField{
				Title: "担当者",
				Value: fmt.Sprintf("<%s>", getUser(ctx, issue.AssignedTo)),
				Short: true,
			},
			&objects.AttachmentField{
				Title: "ステータス",
				Value: issue.Status.Name,
				Short: true,
			},
		},
	}
	slackRESTClient.Chat().PostMessage("#bot_dev").LinkNames(true).Text(msg).Attachment(attachment).Do(ctx)
}

func extractTicketID(s string) int {
	pos := strings.IndexRune(s, '#')
	if pos < 0 {
		return -1
	}
	r := []rune(s)[pos+1:]
	s = string(r)
	for i := range r {
		if !unicode.IsDigit(r[i]) {
			s = string(r[:i])
			break
		}
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		return -1
	}
	return id
}

func getUser(ctx context.Context, idname *redmine.IdName) string {
	id := idname.Name
	if i, ok := userMap[idname.Name]; ok {
		id = i
	}
	ru, err := redmineClient.User(idname.Id)
	if err != nil {
		if id == "channel" {
			return "!" + id
		}
		return id
	}
	if login, ok := userMap[ru.Login]; ok {
		ru.Login = login
	}
	sul, err := slackRESTClient.Users().List().Do(ctx)
	if err != nil {
		return ru.Login
	}
	for _, su := range sul {
		if su.Name == ru.Login {
			return "@" + su.ID
		}
	}
	return ru.Login
}
