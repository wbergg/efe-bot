package tele

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/efe-bot/sbfetch"
	"github.com/wbergg/insultbot/config"
	"github.com/wbergg/telegram"
)

func Run(cfg string, debugTelegram bool, debugStdout bool, telegramTest bool) error {

	// Load config
	config, err := config.LoadConfig(cfg)
	if err != nil {
		log.Error(err)
		panic("Could not load config, check config/config.json")
	}

	channel, err := strconv.ParseInt(config.Telegram.TgChannel, 10, 64)
	if err != nil {
		log.Error(err)
		panic("Could not convert Telegram channel to int64")
	}

	// Initiate telegram
	tg := telegram.New(config.Telegram.TgAPIKey, channel, debugTelegram, debugStdout)
	tg.Init(debugTelegram)

	if telegramTest {
		tg.SendM("DEBUG: efebot test message")
		os.Exit(0)
	}

	// Read messages from Telegram
	updates, err := tg.ReadM()
	if err != nil {
		log.Error(err)
		panic("Cant read from Telegram")
	}

	// Loop
	for update := range updates {

		fmt.Println(update)
		if update.Message == nil { // ignore non-message updates
			continue
		}

		// Debug
		if debugStdout {
			log.Infof("Received message from chat %d [%s]: %s", update.Message.Chat.ID, update.Message.Chat.Type, update.Message.Text)
		}

		if update.Message.IsCommand() {
			// Create switch to search for commands
			switch strings.ToLower(update.Message.Command()) {

			// Insult case
			case "efe":
				message := update.Message.CommandArguments()

				if message == "" {
					// If nothings wa inpuuted, return calling userid
					message = update.Message.From.UserName
					if message == "" {
						message = update.Message.From.FirstName
					}
				}

				// Replace and send
				reply, err := sbfetch.Get(cfg, message)
				if err != nil {
					panic(err)
				}
				tgreply := tgMessageParser(message, reply)

				tg.SendTo(update.Message.Chat.ID, tgreply)

			case "help":
				helpm := `EFEBOT 1.0 - Used to check whether a beer is EFE APPROVED.

				/efe <beer name>

				For example:
				/efe Tuborg Grön`

				tg.SendM(helpm)

			default:
				// Unknown command
				tg.SendM("")
			}
		}
	}

	return err
}

func tgMessageParser(message string, input []sbfetch.Result) string {
	var tgreply string

	posted := make(map[string]bool)

	for _, r := range input {
		if strings.HasPrefix(strings.ToLower(r.NameBold), strings.ToLower(message)) {
			// Generate deduplication key
			key := r.NameBold
			if r.NameThin != "" {
				key += r.NameThin
			}

			// Skip if already posted
			if posted[key] {
				continue
			}
			posted[key] = true

			if r.NameThin != "" {
				if r.Approved {
					tgreply = fmt.Sprintf(tgreply+"\xE2\x9C\x85"+" %s %s, %.1f\n", r.NameBold, r.NameThin, r.Percent)
				} else {
					tgreply = fmt.Sprintf(tgreply+"\xE2\x9D\x8C"+" %s %s, %.1f\n", r.NameBold, r.NameThin, r.Percent)
				}
			} else {
				if r.Approved {
					tgreply = fmt.Sprintf(tgreply+"\xE2\x9C\x85"+" %s, %.1f\n", r.NameBold, r.Percent)
				} else {
					tgreply = fmt.Sprintf(tgreply+"\xE2\x9D\x8C"+" %s, %.1f\n", r.NameBold, r.Percent)
				}
			}

		}
	}

	return tgreply
}
