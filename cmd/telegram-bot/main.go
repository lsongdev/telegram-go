package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lsongdev/telegram-go/telegram"
)

var (
	token = flag.String("token", "", "Telegram Bot Token (required)")
)

func main() {
	flag.Parse()

	if *token == "" {
		log.Fatal("Error: -token flag is required. Please provide your Telegram Bot Token.")
	}

	bot := telegram.NewBot(&telegram.Config{
		Token: *token,
	})

	// Get bot info
	me, err := bot.GetMe()
	if err != nil {
		log.Fatalf("Failed to get bot info: %v", err)
	}
	log.Printf("Bot started: @%s (%s %s)", me.UserName, me.FirstName, me.LastName)

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		cancel()
	}()

	// Start polling for updates
	log.Println("Bot is now polling for messages...")
	bot.StartPolling(ctx, func(update *telegram.Update, err error) {
		if err != nil {
			log.Printf("Error receiving update: %v", err)
			return
		}

		if update.Message == nil {
			return
		}

		msg := update.Message
		log.Printf("[%d] %s: %s", msg.Chat.Id, msg.From.UserName, msg.Text)

		// Handle commands
		if msg.Text != "" {
			switch {
			case msg.Text == "/start":
				reply := fmt.Sprintf("Hello %s! Welcome to the Telegram Bot.", msg.From.FirstName)
				bot.SendMessage(&telegram.MessageRequest{
					ChatId: msg.Chat.Id,
					Text:   reply,
				})

			case msg.Text == "/help":
				helpText := `Available commands:
/start - Start the bot
/help - Show this help message
/echo <text> - Echo your message
/ping - Check bot status
/dice - Roll a dice`
				bot.SendMessage(&telegram.MessageRequest{
					ChatId: msg.Chat.Id,
					Text:   helpText,
				})

			case msg.Text == "/ping":
				bot.SendMessage(&telegram.MessageRequest{
					ChatId: msg.Chat.Id,
					Text:   "Pong! 🏓",
				})

			case len(msg.Text) > 5 && msg.Text[:5] == "/echo":
				echoText := msg.Text[6:]
				if echoText == "" {
					bot.SendMessage(&telegram.MessageRequest{
						ChatId: msg.Chat.Id,
						Text:   "Please provide text to echo. Usage: /echo <text>",
					})
					return
				}
				bot.SendMessage(&telegram.MessageRequest{
					ChatId: msg.Chat.Id,
					Text:   echoText,
				})

			case msg.Text == "/dice":
				bot.SendDice(&telegram.SendDiceRequest{
					ChatId: int(msg.Chat.Id),
					Emoji:  "🎲",
				})

			default:
				// Echo regular messages
				bot.SendMessage(&telegram.MessageRequest{
					ChatId: msg.Chat.Id,
					Text:   fmt.Sprintf("You said: %s", msg.Text),
					ReplyParameters: &telegram.ReplyParameters{
						MessageId: msg.MessageId,
					},
				})
			}
		}
	})
}
