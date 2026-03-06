//go:build example
// +build example

package main

import (
	"context"
	"log"
	"os"

	"github.com/lsongdev/telegram-go/telegram"
)

func main() {
	bot := telegram.NewBot(&telegram.Config{
		Token: os.Getenv("TELEGRAM_BOT_TOKEN"),
	})
	me, err := bot.GetMe()
	if err != nil {
		panic(err)
	}

	log.Println(me)
	ctx, _ := context.WithCancel(context.Background())
	bot.StartPolling(ctx, func(update *telegram.Update, err error) {
		if err != nil {
			log.Println(err)
			return
		}
		log.Println(update.Message.Chat.Id, update.Message.From.UserName, update.Message.Text)
		bot.SendMessageDraft(&telegram.MessageDraftRequest{
			ChatId:  update.Message.Chat.Id,
			DraftId: 1,
			Text:    update.Message.Text,
		})
	})
}
