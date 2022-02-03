package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

func parseWebhookAnswer(resp *http.Response) (string, tgbotapi.Params) {
	var method string
	params := make(tgbotapi.Params)

	var parsed map[string]interface{}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return method, params
	}
	err = json.Unmarshal(respBytes, &parsed)
	if err != nil {
		return method, params
	}
	for key, value := range parsed {
		if key == "method" {
			method = value.(string)
		} else {
			switch value := value.(type) {
			case int64:
				params.AddNonZero64(key, value)
			case float64:
				params.AddNonZeroFloat(key, value)
			case string:
				params.AddNonEmpty(key, value)
			}
		}
	}
	return method, params
}

func main() {
	log.SetOutput(os.Stdout)
	botToken := flag.String("token", "", "Telegram Bot Token")
	webhookUrl := flag.String("url", "", "Bot WebHook url")
	flag.Parse()
	if *botToken == "" || *webhookUrl == "" {
		log.Fatal("Usage: tg-webhook-emulator -token <TELEGRAM_BOT_TOKEN> -url <webhook_url>")
	}
	bot, err := tgbotapi.NewBotAPI(*botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		jsonValue, _ := json.Marshal(update)
		resp, err := http.Post(*webhookUrl, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Error(err)
			continue
		}
		method, params := parseWebhookAnswer(resp)
		apiResp, err := bot.MakeRequest(method, params)

		if err != nil {
			log.Errorf("error: %s, resp: %s", err, apiResp.Result)
			continue
		}
	}
}
