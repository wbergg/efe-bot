package main

import (
	"flag"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/wbergg/efe-bot/config"
	"github.com/wbergg/efe-bot/tele"
)

func main() {
	// Enable bool debug flag
	debugTelegram := flag.Bool("telegram-debug", false, "Turns on debug for telegram")
	debugStdout := flag.Bool("stdout", false, "Turns on stdout rather than sending to telegram")
	telegramTest := flag.Bool("telegram-test", false, "Sends a test message to specified telegram channel")
	configFile := flag.String("config-file", "./config/config.json", "Absolute path for config-file")
	flag.Parse()

	// Load config
	config, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Error(err)
		panic("Could not load config, check config/config.json")
	}
	// DEBUG
	fmt.Println(*debugStdout, *debugTelegram, *telegramTest, config)

	// Run
	//sbfetch.Run(*configFile, *debugTelegram, *debugStdout, *telegramTest)
	tele.Run(*configFile, *debugTelegram, *debugStdout, *telegramTest)

}
