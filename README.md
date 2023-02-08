# telegram-go

> Telegram Bot SDK for Golang

## Usage

```shell
go get -u github.com/song940/telegram-go
```

## Examples

```go
package main

import (
  "context"
  "fmt"
  "log"

  "github.com/song940/telegram-go/telegram"
)

func main() {
  bot := telegram.NewBot(&telegram.Config{
    Token: "-- YOUR_BOT_TOKEN --",
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
    bot.SendMessage(&telegram.MessageRequest{
      ChatId: fmt.Sprintf("%d", update.Message.Chat.Id),
      Text:   update.Message.Text,
    })
  })
}
```

With context
```go

func main() {
  bot := &TelegramBot{}
  ctx, cancel := context.WithCancel(context.Background())
  go bot.StartPolling(ctx, func(update *Update, err error) {
    log.Printf("Received update: %+v\n", update)
  })
  // Do other stuff
  // ...
  cancel()
}
```

## Contributing


## License

This project is licensed under the MIT License.