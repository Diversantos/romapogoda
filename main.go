package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	//	weatherAPIURL = "https://api.weather.yandex.ru/v2/informers"
	weatherAPIURL = "https://api.weather.yandex.ru/v2/forecast"
)

type WeatherResponse struct {
	Fact struct {
		Temperature int    `json:"temp"`
		Condition   string `json:"condition"`
	} `json:"fact"`
}

func getWeather(city string, apiKey string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?lat=%s&lon=%s", weatherAPIURL, "55.75396", "37.620393"), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Yandex-API-Key", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get weather: %s", resp.Status)
	}

	var weatherResponse WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResponse); err != nil {
		return "", err
	}

	return fmt.Sprintf("Температура: %d°C, Условия: %s", weatherResponse.Fact.Temperature, weatherResponse.Fact.Condition), nil
}

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	apiKey := os.Getenv("YANDEX_API_KEY")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message Updates
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Напиши название города, чтобы узнать погоду.")
				bot.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда. Используйте /start.")
				bot.Send(msg)
			}
		} else {
			city := update.Message.Text
			weather, err := getWeather(city, apiKey)
			if err != nil {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка получения погоды: "+err.Error())
				bot.Send(msg)
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, weather)
			bot.Send(msg)
		}
	}
}
