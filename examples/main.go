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

	// Set bot commands menu
	// err := bot.SetMyCommands(&telegram.MyCommandsRequest{
	// 	Scope: &telegram.BotCommandScope{Type: "default"},
	// 	Commands: []*telegram.BotCommand{
	// 		{Command: "start", Description: "Start the bot"},
	// 		{Command: "photo", Description: "Send a photo"},
	// 		{Command: "video", Description: "Send a video"},
	// 		{Command: "document", Description: "Send a document"},
	// 		{Command: "help", Description: "Get help"},
	// 	},
	// })
	// if err != nil {
	// 	log.Println("set commands error:", err)
	// }

	me, err := bot.GetMe()
	if err != nil {
		panic(err)
	}
	log.Printf("%s(@%s)", me.FirstName, me.UserName)

	// Get and print bot commands
	commands, err := bot.GetMyCommands(nil)
	if err != nil {
		log.Println("get commands error:", err)
	} else {
		log.Println("Bot commands:")
		for _, cmd := range commands {
			log.Printf("  /%s - %s", cmd.Command, cmd.Description)
		}
	}

	ctx := context.Background()
	bot.StartPolling(ctx, func(update *telegram.Update, err error) {
		if err != nil {
			log.Println("update error:", err)
		}
		if update.Message != nil {
			handleMessage(bot, update.Message)
		}
		if update.MessageReaction != nil {
			log.Println(update.MessageReaction.OldReaction, "->", update.MessageReaction.NewReaction)
		}
	})
}

func handleMessage(bot *telegram.TelegramBot, message *telegram.Message) {
	log.Printf("%s> %s", message.From.FirstName, message.Text)
	err := bot.SetMessageReaction(telegram.MessageReaction{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
		Reaction:  telegram.NewReaction("❤️"),
	})
	if err != nil {
		log.Panicln(err)
	}

	// Handle commands
	switch message.Text {
	case "/start":
		bot.SendMessage(&telegram.MessageRequest{
			ChatID: message.Chat.ID,
			Text:   "Welcome! Use /help to see available commands.",
		})
	case "/help":
		bot.SendMessage(&telegram.MessageRequest{
			ChatID: message.Chat.ID,
			Text:   "Available commands:\n/photo - Send a photo\n/video - Send a video\n/document - Send a document",
		})
	case "/photo":
		_, err := bot.SendPhoto(&telegram.PhotoRequest{
			ChatID: message.Chat.ID,
			Photo:  "file:///tmp/image.png",
			// Photo:   "https://plus.unsplash.com/premium_photo-1675848495392-6b9a3b962df0",
			Caption: "Here is a photo",
		})
		if err != nil {
			log.Println("upload photo error:", err)
		}
	case "/video":
		// Test SendVideo with local file
		_, err := bot.SendVideo(&telegram.VideoRequest{
			ChatID:  message.Chat.ID,
			Video:   "file://./test_video.mp4",
			Caption: "Here is a 10s video from RTSP stream",
		})
		if err != nil {
			log.Println("upload video error:", err)
		}
	case "/document":
		// Test SendDocument with local file
		_, err := bot.SendDocument(&telegram.DocumentRequest{
			ChatID:   message.Chat.ID,
			Document: "file://./test_video.mp4",
			Caption:  "Here is a document (video file)",
		})
		if err != nil {
			log.Println("upload document error:", err)
		}
	default:
		bot.SendMessage(&telegram.MessageRequest{
			ChatID: message.Chat.ID,
			Text:   message.Text,
			ReplyParameters: &telegram.ReplyParameters{
				ChatID:    message.Chat.ID,
				MessageID: message.MessageID,
			},
		})
	}

}
