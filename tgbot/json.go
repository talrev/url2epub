package tgbot

// Update is a update from telegram webhook.
type Update struct {
	ID       int64          `json:"update_id,omitempty"`
	Message  *Message       `json:"message,omitempty"`
	Callback *CallbackQuery `json:"callback_query,omitempty"`

	// Other not yet supported sub-message types.
	EditedMessage      *Message          `json:"edited_message,omitempty"`
	ChannelPost        *Message          `json:"channel_post,omitempty"`
	EditedChannelPost  *Message          `json:"edited_channel_post,omitempty"`
	Inline             *NotSupportedType `json:"inline,omitempty"`
	ChooseInlineResult *NotSupportedType `json:"choose_inline_result,omitempty"`
	Shipping           *NotSupportedType `json:"shipping_query,omitempty"`
	PreCheckout        *NotSupportedType `json:"pre_checkout_query,omitempty"`
	Poll               *NotSupportedType `json:"poll,omitempty"`
	PollAnswer         *NotSupportedType `json:"poll_answer,omitempty"`
	MyChatMemeber      *NotSupportedType `json:"my_chat_member,omitempty"`
	ChatMemeber        *NotSupportedType `json:"chat_member,omitempty"`
}

// Message is a telegram message.
type Message struct {
	ID   int64  `json:"message_id,omitempty"`
	From User   `json:"from,omitempty"`
	Chat Chat   `json:"chat,omitempty"`
	Date int64  `json:"date,omitempty"`
	Text string `json:"text,omitempty"`

	Entities []MessageEntity `json:"entities,omitempty"`

	ReplyTo *Message `json:"reply_to_message,omitempty"`
}

// User is a telegram user.
type User struct {
	ID        int64  `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Langueage string `json:"language_code,omitempty"`
}

// Chat is a telegram chat.
type Chat struct {
	ID        int64  `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	Type      string `json:"type,omitempty"`
}

// ReplyMessage is a message sent on webhook requests.
type ReplyMessage struct {
	Method  string `json:"method,omitempty"`
	ChatID  int64  `json:"chat_id,omitempty"`
	ReplyTo int64  `json:"reply_to_message_id,omitempty"`
	Text    string `json:"text,omitempty"`

	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// InlineKeyboardMarkup is used to provide single choice replies.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

// InlineKeyboardButton represents a single choice inside InlineKeyboardMarkup.
type InlineKeyboardButton struct {
	Text string `json:"text,omitempty"`
	Data string `json:"callback_data,omitempty"`
}

// CallbackQuery is the callback from InlineKeyboardButton.
type CallbackQuery struct {
	ID      string   `json:"id,omitempty"`
	Data    string   `json:"data,omitempty"`
	Message *Message `json:"message,omitempty"`
}

// AnswerCallbackQuery is the answer to CallbackQuery.
type AnswerCallbackQuery struct {
	ID   string `json:"callback_query_id,omitempty"`
	Text string `json:"text,omitempty"`
}

// MessageEntity represents one special entity in a message (e.g. url)
type MessageEntity struct {
	Type   string `json:"type,omitempty"`
	URL    string `json:"url,omitempty"`
	Offset int64  `json:"offset,omitempty"`
	Length int64  `json:"length,omitempty"`
}

// NotSupportedType defines some of the sub-message types that could happe in
// Update that we do not support yet. It helps us to log to know which
// sub-message type is in the Update message.
//
// These types include: InlineQuery, ChooseInlineResult, ShippingQuery,
// PreCheckoutQuery, Poll, PollAnswer, ChatMemberUpdated.
type NotSupportedType struct {
	ID   string `json:"id,omitempty"`
	Date int64  `json:"date,omitempty"`

	Chat Chat  `json:"chat,omitempty"`
	From *User `json:"from,omitempty"`
	User *User `json:"user,omitempty"`
}
