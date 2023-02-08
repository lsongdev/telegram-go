package main

import (
	"log"

	"github.com/song940/telegram-go/telegram"
)

func main() {
	bot := telegram.NewBot(&telegram.Config{
		Token: "830538223:AAG85jdDekD8RBD_hl4Fih0yeqnJBxn1bWcx",
	})
	me, err := bot.GetMe()
	if err != nil {
		panic(err)
	}
	log.Println(me)
	// resp, err := bot.SendMessage(&telegram.MessageRequest{
	// 	ChatId: "334827694",
	// 	Text:   "Hello Telegram!",
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// log.Println(resp)
	// ctx, _ := context.WithCancel(context.Background())
	// bot.StartPolling(ctx, func(update *telegram.Update, err error) {
	// 	if err != nil {
	// 		log.Println(err)
	// 		return
	// 	}
	// 	log.Println(update.Message.Chat.Id, update.Message.From.UserName, update.Message.Text)
	// 	bot.SendMessage(&telegram.MessageRequest{
	// 		ChatId: fmt.Sprintf("%d", update.Message.Chat.Id),
	// 		Text:   update.Message.Text,
	// 	})
	// })
}
