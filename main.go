package main

import (
	"database/sql"
	"hamed_kocholo_bot/api"
	"hamed_kocholo_bot/models"
	"hamed_kocholo_bot/util"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var userSelections map[int64]*models.UserSelection = make(map[int64]*models.UserSelection)

func main() {
	// Assuming you already have functions to load your config and handle errors:
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(config.TokenBot)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true
	log.Printf("Authorized on bot account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updatesChan := bot.GetUpdatesChan(updateConfig)

	for update := range updatesChan {
		if update.CallbackQuery != nil {
			handleCallbackQuery(update.CallbackQuery, bot, conn)
			continue
		}

		if update.Message != nil && update.Message.IsCommand() {
			switch update.Message.Command() {
			case "send":
				reqType := "set"
				userSelections[update.Message.Chat.ID] = &models.UserSelection{} // reset/init user state
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select a day:")
				msg.ReplyMarkup = api.MakeWeekdayKeyboard(reqType)
				bot.Send(msg)
			case "request":
				reqType := "get"
				userSelections[update.Message.Chat.ID] = &models.UserSelection{}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select a location:")
				msg.ReplyMarkup = api.MakeLocationKeyboard(reqType)
				bot.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "I don't know that command")
				bot.Send(msg)
			}
		} else if update.Message != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "hi"))
		}
	}
}

func handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery, bot *tgbotapi.BotAPI, conn *sql.DB) {
	userSelection := userSelections[callbackQuery.Message.Chat.ID]

	if userSelection == nil {
		userSelection = &models.UserSelection{}
		userSelections[callbackQuery.Message.Chat.ID] = userSelection
	}

	data := callbackQuery.Data
	callbackType := data[:3]
	callbackMenu := data[3:6]
	switch callbackMenu {
	case "day":
		day := data[7:]
		userSelection.SelectedDay = day
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Select a location:")
		msg.ReplyMarkup = api.MakeLocationKeyboard(callbackType)
		bot.Send(msg)
	case "loc":
		location := data[7:]
		userSelection.SelectedLocation = location
		msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Please type the food you'd like:")
		msg.ReplyMarkup = api.MakeFoodKeyboard(callbackType)
		bot.Send(msg)
	case "eat":
		food := data[7:]
		userSelection.SelectedFood = food
		err := api.InsertIntoDatabase(conn, callbackQuery.From.ID, userSelection, callbackType)
		if err != nil {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Error saving your selection.")
			bot.Send(msg)
			log.Printf("Error inserting into database: %v", err)
		} else {
			msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, "Your selections have been saved!")
			bot.Send(msg)
			delete(userSelections, callbackQuery.Message.Chat.ID) // clear user state after successful operation
		}
	}

	// Acknowledge the callback
	callbackConfig := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            "",    // Optional text, not required for just an acknowledgment
		ShowAlert:       false, // Determines if the response should be an alert or a notification at the top of the screen
	}
	if _, err := bot.Request(callbackConfig); err != nil {
		log.Printf("Error responding to callback query: %v", err)
	}
}
