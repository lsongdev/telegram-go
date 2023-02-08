package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Config struct {
	API   string `json:"api"`
	Token string `json:"token"`
}

type TelegramBot struct {
	config *Config
	client *http.Client
}

type TelegramBotResponse struct {
	Ok          bool            `json:"ok"`
	Code        int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
	Result      json.RawMessage `json:"result"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	ID                      int    `json:"id"`
	IsBot                   bool   `json:"is_bot"`
	FirstName               string `json:"first_name"`
	LastName                string `json:"last_name"`
	UserName                string `json:"username"`
	LanguageCode            string `json:"language_code"`
	IsPremium               bool   `json:"is_premium"`
	AddedToAttachmentMenu   bool   `json:"added_to_attachment_menu"`
	CanJoinGroups           bool   `json:"can_join_groups"`
	CanReadAllGroupMessages bool   `json:"can_read_all_group_messages"`
	SupportsInlineQueries   bool   `json:"supports_inline_queries"`
}

type ReplyParameters struct {
	MessageId             int              `json:"message_id"`
	ChatId                int              `json:"chat_id"`
	AllowSendingWithReply bool             `json:"allow_sending_with_reply,omitempty"`
	Quote                 string           `json:"quote,omitempty"`
	QuoteParseMode        string           `json:"quote_parse_mode,omitempty"`
	QuoteEntities         []*MessageEntity `json:"quote_entities,omitempty"`
	QuotePosition         int              `json:"quote_position,omitempty"`
}

type LinkPreviewOptions struct {
	IsDisable        bool   `json:"is_disable,omitempty"`
	URL              string `json:"url,omitempty"`
	PreferSmallMedia bool   `json:"prefer_small_media,omitempty"`
	PreferLargeMedia bool   `json:"prefer_large_media,omitempty"`
	ShowAboveText    bool   `json:"show_above_text,omitempty"`
}

type MessageOrigin struct{}
type ExternalReplyInfo struct{}
type TextQuote struct{}
type Animation struct{}
type PhotoSize struct{}
type Audio struct{}
type Document struct{}
type Sticker struct{}
type Story struct{}
type Video struct{}
type VideoNote struct{}
type Voice struct{}
type Contact struct{}
type Dice struct{}
type Game struct{}
type Poll struct{}
type Venue struct{}
type Location struct{}
type ChatPhoto struct{}
type ReactionType struct{}

// https://core.telegram.org/bots/api#chat
type Chat struct {
	Id                     int             `json:"id"`
	Type                   string          `json:"type"`
	Title                  string          `json:"title"`
	UserName               string          `json:"username"`
	FirstName              string          `json:"first_name"`
	LastName               string          `json:"last_name"`
	IsForum                bool            `json:"is_forum"`
	Photo                  *ChatPhoto      `json:"photo"`
	ActiveUserNames        []string        `json:"active_user_names"`
	AvailableReactions     []*ReactionType `json:"available_reactions"`
	AccentColorId          int             `json:"accent_color"`
	BackgroudCustomEmojiId string          `json:"background_custom_emoji_id"`
	ProfileAccentColorId   int             `json:"profile_accent_color"`
	Bio                    string          `json:"bio"`
	Description            string          `json:"description"`
}

// https://core.telegram.org/bots/api#message
type Message struct {
	MessageId           int                 `json:"message_id"`
	MessageThreadId     int                 `json:"message_thread_id"`
	From                *User               `json:"from"`
	SenderChat          *Chat               `json:"sender_chat"`
	Date                int                 `json:"date"`
	Chat                *Chat               `json:"chat"`
	ForwardOrigin       *MessageOrigin      `json:"forward_origin,omitempty"`
	IsTopicMessage      bool                `json:"is_topic_message"`
	IsAutomaticForward  bool                `json:"is_automatic_forward"`
	ReplyToMessage      *Message            `json:"reply_to_message,omitempty"`
	ExternalReply       *ExternalReplyInfo  `json:"external_reply"`
	Quote               *TextQuote          `json:"quote,omitempty"`
	ViaBot              *User               `json:"via_bot"`
	EditDate            int                 `json:"edit_date,omitempty"`
	HasProtectedContent bool                `json:"has_protected_content,omitempty"`
	MediaGroupId        string              `json:"media_group_id,omitempty"`
	AuthorSignature     string              `json:"author_signature,omitempty"`
	Text                string              `json:"text"`
	Entities            []*MessageEntity    `json:"entities"`
	LinkPreviewOptions  *LinkPreviewOptions `json:"link_preview_options"`
	Animation           *Animation          `json:"animation,omitempty"`
	Audio               *Audio              `json:"audio,omitempty"`
	Document            *Document           `json:"document,omitempty"`
	Photo               []*PhotoSize        `json:"photo,omitempty"`
	Sticker             *Sticker            `json:"sticker,omitempty"`
	Story               *Story              `json:"story,omitempty"`
	Video               *Video              `json:"video,omitempty"`
	VideoNote           *VideoNote          `json:"video_note,omitempty"`
	Voice               *Voice              `json:"voice,omitempty"`
	Caption             *string             `json:"caption,omitempty"`
	CaptionEntities     []*MessageEntity    `json:"caption_entities,omitempty"`
	HasMediaSpoiler     bool                `json:"has_media_spoiler,omitempty"`
	Contact             *Contact            `json:"contact,omitempty"`
	Dice                *Dice               `json:"dice,omitempty"`
	Game                *Game               `json:"game,omitempty"`
	Poll                *Poll               `json:"poll,omitempty"`
	Venue               *Venue              `json:"venue,omitempty"`
	Location            *Location           `json:"location,omitempty"`
	NewChatMembers      []*User             `json:"new_chat_members,omitempty"`
	LeftChatMember      *User               `json:"left_chat_member,omitempty"`
	NewChatTitle        string              `json:"new_chat_title,omitempty"`
	NewChatPhoto        []*PhotoSize        `json:"new_chat_photo,omitempty"`
	DeleteChatPhoto     bool                `json:"delete_chat_photo,omitempty"`
	GroupChatCreated    bool                `json:"group_chat_created,omitempty"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	URL    string `json:"url,omitempty"`
	User   *User  `json:"user,omitempty"`
}

type Update struct {
	UpdateId          int      `json:"update_id"`
	Message           *Message `json:"message,omitempty"`
	EditedMessage     *Message `json:"edited_message,omitempty"`
	ChannelPost       *Message `json:"channel_post,omitempty"`
	EditedChannelPost *Message `json:"edited_channel_post,omitempty"`
}

func NewBot(config *Config) (bot *TelegramBot) {
	bot = &TelegramBot{
		config: config,
		client: http.DefaultClient,
	}
	return
}

func (bot *TelegramBot) Call(method string, params any) (result json.RawMessage, err error) {
	payload, err := json.Marshal(params)
	if err != nil {
		return
	}
	url := "https://api.telegram.org/bot" + bot.config.Token + method
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Add("content-type", "application/json")
	res, err := bot.client.Do(req)
	if err != nil {
		return
	}
	var out TelegramBotResponse
	err = json.NewDecoder(res.Body).Decode(&out)
	if err != nil {
		return
	}
	result = out.Result
	if !out.Ok {
		err = fmt.Errorf("error: %d %s", out.Code, out.Description)
		return
	}
	return
}

// GetMe
// https://core.telegram.org/bots/api#getme
func (bot *TelegramBot) GetMe() (user *User, err error) {
	data, err := bot.Call("/getMe", nil)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &user)
	return
}

type MessageRequest struct {
	ChatId              string              `json:"chat_id"`
	Text                string              `json:"text"`
	MessageThreadId     string              `json:"message_thread_id,omitempty"`
	ParseMode           string              `json:"parse_mode,omitempty"`
	Entities            []*MessageEntity    `json:"entities,omitempty"`
	LinkPreviewOptions  *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	DisableNotification bool                `json:"disable_notification,omitempty"`
	ProtectContent      bool                `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters    `json:"reply_parameters,omitempty"`
	// ReplyMarkup         string             `json:"reply_markup,omitempty"`
}

// SendMessage sends a text message to the specified chat.
// https://core.telegram.org/bots/api#sendmessage
func (bot *TelegramBot) SendMessage(req *MessageRequest) (message *Message, err error) {
	data, err := bot.Call("/sendMessage", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}

// AnswerCallbackQuery sends an answer to a callback query.
// https://core.telegram.org/bots/api#answercallbackquery
func (bot *TelegramBot) AnswerCallbackQuery(callbackQueryId string, text string) error {
	params := map[string]interface{}{
		"callback_query_id": callbackQueryId,
		"text":              text,
	}
	_, err := bot.Call("/answerCallbackQuery", params)
	return err
}

type UpdateRequest struct {
	Offset         int      `json:"offset"`
	Limit          int      `json:"limit"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

// GetUpdates
// https://core.telegram.org/bots/api#getting-updates
func (bot *TelegramBot) GetUpdates(request *UpdateRequest) (updates []*Update, err error) {
	data, err := bot.Call("/getUpdates", request)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &updates)
	return
}
func (bot *TelegramBot) StartPolling(ctx context.Context, updateFunc func(update *Update, err error)) {
	var lastUpdateId int
	for {
		select {
		case <-ctx.Done():
			log.Println("Polling stopped")
			return
		default:
			updates, err := bot.GetUpdates(&UpdateRequest{
				Offset:  lastUpdateId + 1,
				Limit:   100,
				Timeout: 60,
			})
			if err != nil {
				updateFunc(nil, err)
				continue
			}
			for _, update := range updates {
				if update.UpdateId > lastUpdateId {
					lastUpdateId = update.UpdateId
					updateFunc(update, err)
				}
			}
		}
	}
}

type ForwardMessageRequest struct {
	ChatId              int  `json:"chat_id"`
	MessageThreadId     int  `json:"message_thread_id"`
	FromChatId          int  `json:"from_chat_id"`
	DisableNotification bool `json:"disable_notification"`
	ProtectContent      bool `json:"protect_content"`
	MessageId           int  `json:"message_id"`
}

// https://core.telegram.org/bots/api#forwardmessage
func (bot *TelegramBot) ForwardMessage(req *ForwardMessageRequest) (message *Message, err error) {
	data, err := bot.Call("/forwardMessage", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}

type SendLocationRequest struct {
	ChatId               int              `json:"chat_id"`
	MessageThreadId      int              `json:"message_thread_id"`
	Latitute             int              `json:"latitude"`
	Longitude            int              `json:"longitude"`
	HorizontalAccuracy   int              `json:"horizontal_accuracy"`
	LivePeriod           int              `json:"live_period"`
	Heading              int              `json:"heading"`
	ProximityAlertRadius int              `json:"proximity_alert_radius"`
	DisableNotification  bool             `json:"disable_notification"`
	ProtectContent       bool             `json:"protect_content"`
	ReplyParameters      *ReplyParameters `json:"reply_parameters"`
}

// https://core.telegram.org/bots/api#sendlocation
func (bot *TelegramBot) SendLocation(req *SendLocationRequest) (message *Message, err error) {
	data, err := bot.Call("/sendLocation", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}

type SendPollRequest struct {
	ChatId                int              `json:"chat_id"`
	MessageThreadId       int              `json:"message_thread_id"`
	Question              string           `json:"question"`
	Options               []string         `json:"options"`
	IsAnonymous           bool             `json:"is_anonymous"`
	Type                  string           `json:"type"`
	AllowsMultipleAnswers bool             `json:"allows_multiple_answers"`
	CorrectOptionId       int              `json:"correct_option_id"`
	Explanation           string           `json:"explanation"`
	ExplanationParseMode  string           `json:"explanation_parse_mode"`
	ExplanationEntities   []*MessageEntity `json:"explanation_entities"`
	OpenPeriod            int              `json:"open_period"`
	CloseDate             int              `json:"close_date"`
	IsClosed              bool             `json:"is_closed"`
	DisableNotification   bool             `json:"disable_notification"`
	ProtectContent        bool             `json:"protect_content"`
	ReplyParameters       *ReplyParameters `json:"reply_parameters"`
}

// https://core.telegram.org/bots/api#sendpoll
func (bot *TelegramBot) SendPoll(req *SendPollRequest) (message *Message, err error) {
	data, err := bot.Call("/sendPoll", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}

type SendDiceRequest struct {
	ChatId              int              `json:"chat_id"`
	MessageThreadId     int              `json:"message_thread_id"`
	Emoji               string           `json:"emoji"`
	DisableNotification bool             `json:"disable_notification"`
	ProtectContent      bool             `json:"protect_content"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters"`
	// ReplyMarkup         *ReplyMarkup     `json:"reply_markup"`
}

// https://core.telegram.org/bots/api#senddice
func (bot *TelegramBot) SendDice(req *SendDiceRequest) (message *Message, err error) {
	data, err := bot.Call("/sendDice", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}
