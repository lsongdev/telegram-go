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
	log.Printf("%s(@%s)", me.FirstName, me.UserName)

	ctx := context.Background()
	bot.StartPolling(ctx, func(update *telegram.Update, err error) {
		if err != nil {
			log.Fatal(err)
		}
		if update.Message != nil {
			log.Printf("%s> %s", update.Message.From.FirstName, update.Message.Text)

			err = bot.SetMessageReaction(telegram.MessageReaction{
				ChatID:    update.Message.Chat.ID,
				MessageID: update.Message.MessageID,
				Reaction:  telegram.NewReaction("❤️"),
			})
			if err != nil {
				log.Panicln(err)
			}
		}
		if update.MessageReaction != nil {
			log.Println(update.MessageReaction.OldReaction, "->", update.MessageReaction.NewReaction)
		}

	})
}
