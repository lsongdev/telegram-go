package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	API   string `json:"api"`
	Token string `json:"token"`
}

type TelegramBot struct {
	config          *Config
	client          *http.Client
	IncomingMessage chan *Update
}

type TelegramBotResponse struct {
	Ok          bool            `json:"ok"`
	Code        int             `json:"error_code,omitempty"`
	Description string          `json:"description,omitempty"`
	Result      json.RawMessage `json:"result"`
}

// https://core.telegram.org/bots/api#user
type User struct {
	ID                      int64  `json:"id"`
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
	// Unique identifier for this chat.
	// This number may have more than 32 significant bits and some programming languages may have difficulty/silent defects in interpreting it.
	// But it has at most 52 significant bits, so a signed 64-bit integer or double-precision float type are safe for storing this identifier.
	ID                     int64           `json:"id"`
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
	MessageID           int64               `json:"message_id"`
	MessageThreadID     int64               `json:"message_thread_id"`
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

func NewBot(config *Config) (bot *TelegramBot) {
	if config.Token == "" {
		log.Fatalln("token is empty")
	}
	bot = &TelegramBot{
		config: config,
		client: http.DefaultClient,
	}
	return
}

func (bot *TelegramBot) requestJson(path string, params any) (result json.RawMessage, err error) {
	body := &bytes.Buffer{}
	err = json.NewEncoder(body).Encode(params)
	if err != nil {
		return
	}
	return bot.request(path, body, map[string]string{
		"Content-Type": "application/json",
	})
}

func (bot *TelegramBot) requestForm(path string, form map[string]any) (result json.RawMessage, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for fieldName, value := range form {
		f, ok := value.(*os.File)
		if ok {
			part, err := writer.CreateFormFile(fieldName, filepath.Base(f.Name()))
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(part, f)
			if err != nil {
				return nil, err
			}
			err = f.Close()
			if err != nil {
				return nil, err
			}
		} else {
			err = writer.WriteField(fieldName, fmt.Sprintf("%v", value))
			if err != nil {
				return nil, err
			}
		}
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return bot.request(path, body, map[string]string{
		"Content-Type": writer.FormDataContentType(),
	})
}

// @docs https://core.telegram.org/bots/api#making-requests
func (bot *TelegramBot) request(path string, body io.Reader, headers map[string]string) (result json.RawMessage, err error) {
	url := "https://api.telegram.org/bot" + bot.config.Token + path
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return
	}
	for name, value := range headers {
		req.Header.Add(name, value)
	}
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

// CallMethod is a generic method to call any Telegram Bot API method.
// - method: the API method name (e.g., "getMe", "sendMessage")
// - params: request parameters (struct or map[string]any)
// - out: pointer to result struct to unmarshal the response
// Returns error if the API call fails or returns a non-success response.
func (bot *TelegramBot) CallMethod(method string, params any, out any) (err error) {
	path := fmt.Sprintf("/%s", method)
	var result json.RawMessage
	form, ok := params.(map[string]any)
	if ok {
		result, err = bot.requestForm(path, form)
	} else {
		result, err = bot.requestJson(path, params)
	}
	if err != nil {
		return
	}
	if out != nil {
		err = json.Unmarshal(result, out)
		return
	}
	// For methods returning boolean true
	if string(result) != "true" {
		err = fmt.Errorf("error: %s", string(result))
		return
	}
	return nil
}

// GetMe
// https://core.telegram.org/bots/api#getme
func (bot *TelegramBot) GetMe() (user *User, err error) {
	err = bot.CallMethod("getMe", nil, &user)
	return
}

type UpdateRequest struct {
	Offset         int      `json:"offset"`
	Limit          int      `json:"limit"`
	Timeout        int      `json:"timeout"`
	AllowedUpdates []string `json:"allowed_updates"`
}

type Update struct {
	UpdateId          int      `json:"update_id"`
	Message           *Message `json:"message,omitempty"`
	EditedMessage     *Message `json:"edited_message,omitempty"`
	ChannelPost       *Message `json:"channel_post,omitempty"`
	EditedChannelPost *Message `json:"edited_channel_post,omitempty"`
	// business_connection
	// business_message
	// edited_business_message
	// deleted_business_messages
	MessageReaction      *MessageReactionUpdated      `json:"message_reaction,omitempty"`
	MessageReactionCount *MessageReactionCountUpdated `json:"message_reaction_count,omitempty"`
	// inline_query
	// chosen_inline_result
	// callback_query
	// shipping_query
	// pre_checkout_query
	// purchased_paid_media
	// poll
	// poll_answer
	// my_chat_member
	// chat_member
	// chat_join_request
	// chat_boost
	// removed_chat_boost
}

// https://core.telegram.org/bots/api#messagereactionupdated
type MessageReactionUpdated struct {
	MessageID   int64      `json:"message_id"`
	Chat        Chat       `json:"chat"`
	User        User       `json:"user,omitempty"`
	ActorChat   Chat       `json:"actor_chat,omitempty"`
	Date        int64      `json:"date"`
	OldReaction []Reaction `json:"old_reaction"`
	NewReaction []Reaction `json:"new_reaction"`
}

// https://core.telegram.org/bots/api#messagereactionupdated
type Reaction struct {
	Type          string `json:"type"` // "emoji" | "custom_emoji" | "paid"
	Emoji         string `json:"emoji"`
	CustomEmojiID string `json:"custom_emoji_id"`
}

// https://core.telegram.org/bots/api#messagereactionupdated
type MessageReactionCountUpdated struct {
	MessageID int64           `json:"message_id"`
	Chat      Chat            `json:"chat"`
	Date      int64           `json:"date"`
	Reactions []ReactionCount `json:"reactions"`
}

type ReactionCount struct {
	Type       string `json:"type"`
	TotalCount int    `json:"total_count"`
}

// GetUpdates
// https://core.telegram.org/bots/api#getting-updates
func (bot *TelegramBot) GetUpdates(request *UpdateRequest) (updates []*Update, err error) {
	err = bot.CallMethod("getUpdates", request, &updates)
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

func (bot *TelegramBot) Start(ctx context.Context) {
	bot.StartPolling(ctx, func(update *Update, err error) {
		if err != nil {
			log.Println(err)
			return
		}
		bot.IncomingMessage <- update
	})
}

type MessageRequest struct {
	// business_connection_id
	ChatID          any   `json:"chat_id"`
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// direct_messages_topic_id
	Text                string              `json:"text"`
	ParseMode           string              `json:"parse_mode,omitempty"`
	Entities            []*MessageEntity    `json:"entities,omitempty"`
	LinkPreviewOptions  *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	DisableNotification bool                `json:"disable_notification,omitempty"`
	ProtectContent      bool                `json:"protect_content,omitempty"`
	// allow_paid_broadcast bool
	// message_effect_id string
	// suggested_post_parameters
	ReplyParameters *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup     any              `json:"reply_markup,omitempty"`
}

// @docs https://core.telegram.org/bots/api#replyparameters
type ReplyParameters struct {
	MessageID                int64            `json:"message_id"`
	ChatID                   any              `json:"chat_id,omitempty"`
	AllowSendingWithoutReply bool             `json:"allow_sending_without_reply,omitempty"`
	Quote                    string           `json:"quote,omitempty"`
	QuoteParseMode           string           `json:"quote_parse_mode,omitempty"`
	QuoteEntities            []*MessageEntity `json:"quote_entities,omitempty"`
	QuotePosition            int              `json:"quote_position,omitempty"`
	ChecklistTaskID          int              `json:"checklist_task_id,omitempty"`
}

type InlineKeyboardMarkup struct{}
type ReplyKeyboardMarkup struct{}
type ReplyKeyboardRemove struct{}
type ForceReply struct{}

// SendMessage sends a text message to the specified chat.
// https://core.telegram.org/bots/api#sendmessage
func (bot *TelegramBot) SendMessage(message *MessageRequest) (result *Message, err error) {
	err = bot.CallMethod("sendMessage", message, &result)
	return
}

type ForwardMessageRequest struct {
	ChatID              int  `json:"chat_id"`
	MessageThreadID     int  `json:"message_thread_id"`
	FromChatId          int  `json:"from_chat_id"`
	DisableNotification bool `json:"disable_notification"`
	ProtectContent      bool `json:"protect_content"`
	MessageID           int  `json:"message_id"`
}

// https://core.telegram.org/bots/api#forwardmessage
func (bot *TelegramBot) ForwardMessage(req *ForwardMessageRequest) (result *Message, err error) {
	err = bot.CallMethod("forwardMessage", req, &result)
	return
}

type SendLocationRequest struct {
	// business_connection_id
	ChatID          any   `json:"chat_id"`
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// direct_messages_topic_id
	Latitute             float32 `json:"latitude"`
	Longitude            float32 `json:"longitude"`
	HorizontalAccuracy   int     `json:"horizontal_accuracy,omitempty"`
	LivePeriod           int     `json:"live_period,omitempty"`
	Heading              int     `json:"heading,omitempty"`
	ProximityAlertRadius int     `json:"proximity_alert_radius,omitempty"`
	DisableNotification  bool    `json:"disable_notification,omitempty"`
	ProtectContent       bool    `json:"protect_content,omitempty"`
	// allow_paid_broadcast
	// message_effect_id
	// suggested_post_parameters
	ReplyParameters *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup     any              `json:"reply_markup,omitempty"`
}

// https://core.telegram.org/bots/api#sendlocation
func (bot *TelegramBot) SendLocation(req *SendLocationRequest) (result *Message, err error) {
	err = bot.CallMethod("sendLocation", req, &result)
	return
}

type InputPollOption struct {
}

type SendPollRequest struct {
	// business_connection_id
	ChatID                any               `json:"chat_id"`
	MessageThreadID       int               `json:"message_thread_id,omitempty"`
	Question              string            `json:"question"`
	QuestionParseMode     string            `json:"question_parse_mode,omitempty"`
	QuestionEntities      []*MessageEntity  `json:"question_entities,omitempty"`
	Options               []InputPollOption `json:"options"`
	IsAnonymous           bool              `json:"is_anonymous,omitempty"`
	Type                  string            `json:"type,omitempty"`
	AllowsMultipleAnswers bool              `json:"allows_multiple_answers,omitempty"`
	CorrectOptionID       int               `json:"correct_option_id,omitempty"`
	Explanation           string            `json:"explanation,omitempty"`
	ExplanationParseMode  string            `json:"explanation_parse_mode,omitempty"`
	ExplanationEntities   []*MessageEntity  `json:"explanation_entities,omitempty"`
	OpenPeriod            int               `json:"open_period,omitempty"`
	CloseDate             int               `json:"close_date,omitempty"`
	IsClosed              bool              `json:"is_closed,omitempty"`
	DisableNotification   bool              `json:"disable_notification,omitempty"`
	ProtectContent        bool              `json:"protect_content,omitempty"`
	// allow_paid_broadcast
	// message_effect_id
	ReplyParameters *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup     any              `json:"reply_markup,omitempty"`
}

// https://core.telegram.org/bots/api#sendpoll
func (bot *TelegramBot) SendPoll(req *SendPollRequest) (result *Message, err error) {
	err = bot.CallMethod("sendPoll", req, &result)
	return
}

type SendDiceRequest struct {
	// business_connection_id
	ChatID          any   `json:"chat_id"`
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// direct_messages_topic_id
	Emoji               string           `json:"emoji,omitempty"`
	DisableNotification bool             `json:"disable_notification"`
	ProtectContent      bool             `json:"protect_content"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters"`
	ReplyMarkup         any              `json:"reply_markup"`
}

// https://core.telegram.org/bots/api#senddice
func (bot *TelegramBot) SendDice(req *SendDiceRequest) (result *Message, err error) {
	err = bot.CallMethod("sendDice", req, &result)
	return
}

type EditMessageTextRequest struct {
	ChatID             any                 `json:"chat_id,omitempty"`
	MessageID          int64               `json:"message_id,omitempty"`
	InlineMessageID    string              `json:"inline_message_id,omitempty"`
	Text               string              `json:"text"`
	ParseMode          string              `json:"parse_mode,omitempty"`
	Entities           []*MessageEntity    `json:"entities,omitempty"`
	LinkPreviewOptions *LinkPreviewOptions `json:"link_preview_options,omitempty"`
	ReplyMarkup        any                 `json:"reply_markup,omitempty"`
}

// https://core.telegram.org/bots/api#editmessagetext
func (bot *TelegramBot) EditMessageText(req *EditMessageTextRequest) (message *Message, err error) {
	data, err := bot.requestJson("/editMessageText", req)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &message)
	return
}

type MessageDraftRequest struct {
	ChatID          int64            `json:"chat_id"`
	MessageThreadID int64            `json:"message_thread_id,omitempty"`
	DraftID         int64            `json:"draft_id"`
	Text            string           `json:"text"`
	ParseMode       string           `json:"parse_mode,omitempty"`
	Entities        []*MessageEntity `json:"entities,omitempty"`
}

// Use this method to stream a partial message to a user while the message is being generated. Returns True on success.
// https://core.telegram.org/bots/api#sendmessagedraft
func (bot *TelegramBot) SendMessageDraft(req *MessageDraftRequest) error {
	return bot.CallMethod("sendMessageDraft", req, nil)
}

type ChatAction struct {
	// business_connection_id
	ChatID          any    `json:"chat_id"`
	MessageThreadID int64  `json:"message_thread_id,omitempty"`
	Action          string `json:"action"`
}

// SendChatAction sends a chat action to show status (typing, upload_photo, etc.)
// https://core.telegram.org/bots/api#sendchataction
func (bot *TelegramBot) SendChatAction(action *ChatAction) error {
	_, err := bot.requestJson("/sendChatAction", action)
	return err
}

type MessageReaction struct {
	// Unique identifier for the target chat or username of the target channel (in the format @channelusername)
	ChatID    any        `json:"chat_id"`
	MessageID int64      `json:"message_id"`
	Reaction  []Reaction `json:"reaction,omitempty"`
	IsBig     bool       `json:"is_big,omitempty"`
}

// Use this method to change the chosen reactions on a message.
// Service messages of some types can't be reacted to.
// Automatically forwarded messages from a channel to its discussion group have the same available reactions as messages in the channel.
// Bots can't use paid reactions. Returns True on success.
// @docs https://core.telegram.org/bots/api#setmessagereaction
func (bot *TelegramBot) SetMessageReaction(reaction MessageReaction) error {
	_, err := bot.requestJson("/setMessageReaction", reaction)
	return err
}

func NewReaction(emojis ...string) (reactions []Reaction) {
	for _, emoji := range emojis {
		reaction := Reaction{
			Type:  "emoji",
			Emoji: emoji,
		}
		reactions = append(reactions, reaction)
	}
	return
}

type PhotoRequest struct {
	// business_connection_id
	ChatID              any              `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Photo               string           `json:"photo"` // file_id, URL, or "attach://file_name" for file upload
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	CaptionEntities     []*MessageEntity `json:"caption_entities,omitempty"`
	HasSpoiler          bool             `json:"has_spoiler,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

// handleFileField handles file upload for Telegram bot API fields.
// If the fieldValue starts with "file://", it opens the local file and adds it to the form.
// Returns the form map, the opened file (if any), and any error.
func prepareForm(params any, fieldName string) (map[string]any, *os.File, error) {
	form := ToFormValues(params)
	result := make(map[string]any)
	for k, v := range form {
		result[k] = v
	}
	fieldValue := form[fieldName]
	if !strings.HasPrefix(fieldValue, "file://") {
		return result, nil, nil
	}
	filePath := strings.TrimPrefix(fieldValue, "file://")
	f, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	fileName := filepath.Base(f.Name())
	result[fileName] = f
	result[fieldName] = fmt.Sprintf("attach://%s", fileName)
	return result, f, nil
}

// SendPhoto sends a photo to the specified chat.
// Photo can be a file_id, URL, or "attach://file_name" for file upload.
// When using "attach://file_name", set File field to the local file path.
// https://core.telegram.org/bots/api#sendphoto
func (bot *TelegramBot) SendPhoto(req *PhotoRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "photo")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendPhoto", form, &result)
	return
}

type VideoRequest struct {
	// business_connection_id
	ChatID              any              `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Video               string           `json:"video"` // file_id, URL, or "attach://file_name" for file upload
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	CaptionEntities     []*MessageEntity `json:"caption_entities,omitempty"`
	HasSpoiler          bool             `json:"has_spoiler,omitempty"`
	Duration            int              `json:"duration,omitempty"`
	Width               int              `json:"width,omitempty"`
	Height              int              `json:"height,omitempty"`
	Thumbnail           string           `json:"thumbnail,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

// SendVideo sends a video to the specified chat.
// Video can be a file_id, URL, or "attach://file_name" for file upload.
// https://core.telegram.org/bots/api#sendvideo
func (bot *TelegramBot) SendVideo(req *VideoRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "video")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendVideo", form, &result)
	return
}

type DocumentRequest struct {
	// business_connection_id
	ChatID                      any              `json:"chat_id"`
	MessageThreadID             int64            `json:"message_thread_id,omitempty"`
	Document                    string           `json:"document"` // file_id, URL, or "attach://file_name" for file upload
	Caption                     string           `json:"caption,omitempty"`
	ParseMode                   string           `json:"parse_mode,omitempty"`
	CaptionEntities             []*MessageEntity `json:"caption_entities,omitempty"`
	DisableContentTypeDetection bool             `json:"disable_content_type_detection,omitempty"`
	Thumbnail                   string           `json:"thumbnail,omitempty"`
	DisableNotification         bool             `json:"disable_notification,omitempty"`
	ProtectContent              bool             `json:"protect_content,omitempty"`
	ReplyParameters             *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup                 any              `json:"reply_markup,omitempty"`
}

// SendDocument sends a document to the specified chat.
// Document can be a file_id, URL, or "attach://file_name" for file upload.
// https://core.telegram.org/bots/api#senddocument
func (bot *TelegramBot) SendDocument(req *DocumentRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "document")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendDocument", form, &result)
	return
}

type AudioRequest struct {
	// business_connection_id
	ChatID          any   `json:"chat_id"`
	MessageThreadID int64 `json:"message_thread_id,omitempty"`
	// direct_messages_topic_id
	Audio               string           `json:"audio"` // file_id, URL, or "attach://file_name" for file upload
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	CaptionEntities     []*MessageEntity `json:"caption_entities,omitempty"`
	Duration            int              `json:"duration,omitempty"`
	Performer           string           `json:"performer,omitempty"`
	Title               string           `json:"title,omitempty"`
	Thumbnail           string           `json:"thumbnail,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	// allow_paid_broadcast
	// message_effect_id
	// suggested_post_parameters
	ProtectContent  bool             `json:"protect_content,omitempty"`
	ReplyParameters *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup     any              `json:"reply_markup,omitempty"`
}

// SendAudio sends an audio file to the specified chat.
// Audio can be a file_id, URL, or "attach://file_name" for file upload.
// https://core.telegram.org/bots/api#sendaudio
func (bot *TelegramBot) SendAudio(req *AudioRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "audio")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendAudio", form, &result)
	return
}

type VoiceRequest struct {
	// business_connection_id
	ChatID              any              `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Voice               string           `json:"voice"` // file_id, URL, or "attach://file_name" for file upload
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	CaptionEntities     []*MessageEntity `json:"caption_entities,omitempty"`
	Duration            int              `json:"duration,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

// SendVoice sends a voice message to the specified chat.
// Voice can be a file_id, URL, or "attach://file_name" for file upload.
// https://core.telegram.org/bots/api#sendvoice
func (bot *TelegramBot) SendVoice(req *VoiceRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "voice")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendVoice", form, &result)
	return
}

type AnimationRequest struct {
	// business_connection_id
	ChatID              any              `json:"chat_id"`
	MessageThreadID     int64            `json:"message_thread_id,omitempty"`
	Animation           string           `json:"animation"` // file_id, URL, or "attach://file_name" for file upload
	Caption             string           `json:"caption,omitempty"`
	ParseMode           string           `json:"parse_mode,omitempty"`
	CaptionEntities     []*MessageEntity `json:"caption_entities,omitempty"`
	HasSpoiler          bool             `json:"has_spoiler,omitempty"`
	Duration            int              `json:"duration,omitempty"`
	Width               int              `json:"width,omitempty"`
	Height              int              `json:"height,omitempty"`
	Thumbnail           string           `json:"thumbnail,omitempty"`
	DisableNotification bool             `json:"disable_notification,omitempty"`
	ProtectContent      bool             `json:"protect_content,omitempty"`
	ReplyParameters     *ReplyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup         any              `json:"reply_markup,omitempty"`
}

// SendAnimation sends an animation file (GIF or H.264 MP4) to the specified chat.
// Animation can be a file_id, URL, or "attach://file_name" for file upload.
// https://core.telegram.org/bots/api#sendanimation
func (bot *TelegramBot) SendAnimation(req *AnimationRequest) (result *Message, err error) {
	form, f, err := prepareForm(req, "animation")
	if err != nil {
		return nil, err
	}
	if f != nil {
		defer f.Close()
	}
	err = bot.CallMethod("sendAnimation", form, &result)
	return
}

type ChatMenuButton struct {
	ChatID     int64      `json:"chat_id,omitempty"`
	MenuButton MenuButton `json:"menu_button,omitempty"`
}

type MenuButton struct {
	Type string `json:"type"` // "commands" | "web_app" | "default"
}

func (bot *TelegramBot) SetChatMenuButton(button *ChatMenuButton) error {
	return bot.CallMethod("setChatMenuButton", button, nil)
}

// BotCommand represents a bot command.
// @docs https://core.telegram.org/bots/api#botcommand
type BotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// BotCommandScope represents the scope of bot commands.
// @docs https://core.telegram.org/bots/api#botcommandscope
type BotCommandScope struct {
	Type string `json:"type"` // "default" | "all_private_chats" | "all_group_chats" | "all_chat_administrators" | "chat" | "chat_administrators" | "chat_member"
}

// CommandRequest is the request for setting bot commands.
// @docs https://core.telegram.org/bots/api#setmycommands
type MyCommandsRequest struct {
	Commands     []*BotCommand    `json:"commands,omitempty"`
	Scope        *BotCommandScope `json:"scope,omitempty"`
	LanguageCode string           `json:"language_code,omitempty"`
}

// SetMyCommands sets the list of commands for the bot.
// Use scope to set commands for different chat types.
// Example:
//
//	bot.SetMyCommands(&CommandRequest{
//		Commands: []BotCommand{
//			{Command: "start", Description: "Start the bot"},
//			{Command: "help", Description: "Get help"},
//		},
//		Scope: &BotCommandScope{Type: "default"},
//	})
//
// @docs https://core.telegram.org/bots/api#setmycommands
func (bot *TelegramBot) SetMyCommands(req *MyCommandsRequest) error {
	return bot.CallMethod("setMyCommands", req, nil)
}

// GetMyCommands gets the list of commands for the bot.
// @docs https://core.telegram.org/bots/api#getmycommands
func (bot *TelegramBot) GetMyCommands(req *MyCommandsRequest) (commands []BotCommand, err error) {
	err = bot.CallMethod("getMyCommands", req, &commands)
	return
}

// DeleteMyCommands deletes the list of commands for the bot.
// @docs https://core.telegram.org/bots/api#deletemycommands
func (bot *TelegramBot) DeleteMyCommands(req *MyCommandsRequest) error {
	return bot.CallMethod("deleteMyCommands", req, nil)
}
