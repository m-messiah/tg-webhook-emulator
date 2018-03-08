package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func parseWebhookAnswer(resp *http.Response) (string, url.Values) {
	v := url.Values{}
	var method string
	for i := 0; i < 3; i++ {
		if i == 0 {
			var parsed map[string]interface{}
			bytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			err = json.Unmarshal(bytes, &parsed)
			if err != nil {
				continue
			}
			for key, value := range parsed {
				if key == "method" {
					method = value.(string)
				} else {
					switch value.(type) {
					case int64:
						v.Add(key, strconv.FormatInt(value.(int64), 10))
					case float64:
						v.Add(key, strconv.FormatFloat(value.(float64), 'f', 0, 64))
					case string:
						v.Add(key, value.(string))
					}
				}
			}
			return method, v
		} else {
			// Implement other values methods here
		}
	}
	return method, v
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

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		jsonValue, _ := json.Marshal(update)
		resp, err := http.Post(*webhookUrl, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			log.Error(err)
			continue
		}
		method, params := parseWebhookAnswer(resp)
		_, err = bot.MakeRequest(method, params)

		if err != nil {
			log.Error(err)
			continue
		}
	}
}
