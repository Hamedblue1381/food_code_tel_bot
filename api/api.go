package api

import (
	"database/sql"
	"fmt"
	"hamed_kocholo_bot/models"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	_ "github.com/lib/pq"
)

func InsertIntoDatabase(conn *sql.DB, userID int64, userSelection *models.UserSelection, reqType string) error {
	switch reqType {
	case "set":
		stmt, err := conn.Prepare("INSERT INTO codes (chat_id, selected_day, selected_location, food_name) VALUES ($1, $2, $3, $4) RETURNING id")
		if err != nil {
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(userID, userSelection.SelectedDay, userSelection.SelectedLocation, userSelection.SelectedFood)
		return err
	case "get":
		stmt, err := conn.Prepare("INSERT INTO requests (chat_id, selected_day, selected_location, food_name) VALUES ($1, $2, $3, $4) RETURNING id")
		if err != nil {
			return err
		}
		defer stmt.Close()
		value := models.WeekdaysMap[time.Now().Weekday().String()]

		_, err = stmt.Exec(userID, value, userSelection.SelectedLocation, userSelection.SelectedFood)
		return err

	default:
		return fmt.Errorf("unknown request type: %s", reqType)
	}

}

func MakeWeekdayKeyboard(reqType string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, weekday := range models.WeekdaysMap {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(weekday, fmt.Sprintf(reqType+"day_%v", weekday)),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
func MakeFoodKeyboard(reqType string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, food := range models.FoodList {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(food, fmt.Sprintf(reqType+"eat_%v", food)),
		))
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func MakeLocationKeyboard(reqType string) tgbotapi.InlineKeyboardMarkup {
	buttons := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("خوابگاه", reqType+"loc_خوابگاه"),
		tgbotapi.NewInlineKeyboardButtonData("سلف", reqType+"loc_سلف"),
	}
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttons...))
}
