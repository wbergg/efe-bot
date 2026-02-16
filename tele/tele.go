package tele

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/efe-bot/bsfetch"
	"github.com/wbergg/efe-bot/sbfetch"
	"github.com/wbergg/efe-bot/config"
	"github.com/wbergg/telegram"
)

func Run(cfg string, debugTelegram bool, debugStdout bool, telegramTest bool) error {

	// Ratelimit variables
	var sbfetchMutex sync.Mutex
	var lastFetchTime time.Time
	var rateLimitDelay = 5 * time.Second

	// Load config
	config, err := config.LoadConfig(cfg)
	if err != nil {
		return fmt.Errorf("could not load config: %w", err)
	}

	// TG channel
	channel, err := strconv.ParseInt(config.Telegram.TgChannel, 10, 64)
	if err != nil {
		return fmt.Errorf("could not convert Telegram channel to int64: %w", err)
	}

	// Initiate telegram
	tg := telegram.New(config.Telegram.TgAPIKey, channel, debugTelegram, debugStdout)
	tg.Init(debugTelegram)

	// TG test
	if telegramTest {
		tg.SendM("DEBUG: efebot test message")
		os.Exit(0)
	}

	// Read messages from Telegram
	updates, err := tg.ReadM()
	if err != nil {
		return fmt.Errorf("cant read from Telegram: %w", err)
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

				// Lock
				sbfetchMutex.Lock()
				time_now := time.Now()
				if time_now.Sub(lastFetchTime) < rateLimitDelay {
					sbfetchMutex.Unlock()
					tg.SendTo(update.Message.Chat.ID, "Throttled - Please wait before trying again.")
					break
				}
				lastFetchTime = time_now
				// Unlock
				sbfetchMutex.Unlock()

				// Fetch from both APIs in parallel
				var wg sync.WaitGroup
				var sbReply []sbfetch.Result
				var bsReply []bsfetch.Result
				var sbErr, bsErr error

				wg.Add(2)
				go func() {
					defer wg.Done()
					sbReply, sbErr = sbfetch.Get(config, message)
				}()
				go func() {
					defer wg.Done()
					bsReply, bsErr = bsfetch.Get(config, message)
				}()
				wg.Wait()

				if sbErr != nil {
					log.Error("Error fetching from Systembolaget: ", sbErr)
				}
				if bsErr != nil {
					log.Error("Error fetching from Bordershop: ", bsErr)
				}

				// Combine results from both APIs
				var combinedResults []sbfetch.Result
				combinedResults = append(combinedResults, sbReply...)

				// Convert bsfetch.Result to sbfetch.Result and append
				for _, bsResult := range bsReply {
					combinedResults = append(combinedResults, sbfetch.Result{
						NameBold: bsResult.NameBold,
						NameThin: bsResult.NameThin,
						Percent:  bsResult.Percent,
						Approved: bsResult.Approved,
					})
				}

				// Check if we got any results at all
				if len(combinedResults) == 0 {
					tg.SendTo(update.Message.Chat.ID, "Sorry, no results found or there was an error searching. Please try again later.")
					break
				}

				// Parse combined reply
				tgreply := tgMessageParser(message, combinedResults)

				// Send message
				tg.SendTo(update.Message.Chat.ID, tgreply)

			case "help":
				// Help message
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
		if strings.Contains(strings.ToLower(r.NameBold), strings.ToLower(message)) {
			// Dupliceate check
			key := r.NameBold
			if r.NameThin != "" {
				key += r.NameThin
			}

			// Stop loop if posted
			if posted[key] {
				continue
			}

			posted[key] = true

			// Check if NameBold already contains percentage (from bordershop)
			hasPercent := strings.Contains(r.NameBold, "%")

			if r.NameThin != "" {
				if hasPercent {
					// Bordershop format: name already includes percentage
					if r.Approved {
						tgreply += fmt.Sprintf("\xE2\x9C\x85"+" %s %s (source Bordershop)\n", r.NameBold, r.NameThin)
					} else {
						tgreply += fmt.Sprintf("\xE2\x9D\x8C"+" %s %s (source Bordershop)\n", r.NameBold, r.NameThin)
					}
				} else {
					// Systembolaget format: need to append percentage
					if r.Approved {
						tgreply += fmt.Sprintf("\xE2\x9C\x85"+" %s %s %.1f%% (source Systembolaget)\n", r.NameBold, r.NameThin, r.Percent)
					} else {
						tgreply += fmt.Sprintf("\xE2\x9D\x8C"+" %s %s %.1f%% (source Systembolaget)\n", r.NameBold, r.NameThin, r.Percent)
					}
				}
			} else {
				if hasPercent {
					// Bordershop format: name already includes percentage
					if r.Approved {
						tgreply += fmt.Sprintf("\xE2\x9C\x85"+" %s (source Bordershop)\n", r.NameBold)
					} else {
						tgreply += fmt.Sprintf("\xE2\x9D\x8C"+" %s (source Bordershop)\n", r.NameBold)
					}
				} else {
					// Systembolaget format: need to append percentage
					if r.Approved {
						tgreply += fmt.Sprintf("\xE2\x9C\x85"+" %s %.1f%% (source Systembolaget)\n", r.NameBold, r.Percent)
					} else {
						tgreply += fmt.Sprintf("\xE2\x9D\x8C"+" %s %.1f%% (source Systembolaget)\n", r.NameBold, r.Percent)
					}
				}
			}

		}
	}

	return tgreply
}
