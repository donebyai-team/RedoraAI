package models

type CallMessage struct {
	UserMessage   *UserMessage
	SystemMessage *SystemMessage
	BotMessage    *BotMessage
}

type UserMessage struct {
	// The role of the user in the conversation.
	Role string `json:"role" url:"role"`
	// The message content from the user.
	Message string `json:"message" url:"message"`
	// The timestamp when the message was sent.
	Time float64 `json:"time" url:"time"`
	// The timestamp when the message ended.
	EndTime float64 `json:"endTime" url:"endTime"`
	// The number of seconds from the start of the conversation.
	SecondsFromStart float64 `json:"secondsFromStart" url:"secondsFromStart"`
	// The duration of the message in seconds.
	Duration *float64 `json:"duration,omitempty" url:"duration,omitempty"`
}

type SystemMessage struct {
	// The role of the system in the conversation.
	Role string `json:"role" url:"role"`
	// The message content from the system.
	Message string `json:"message" url:"message"`
	// The timestamp when the message was sent.
	Time float64 `json:"time" url:"time"`
	// The number of seconds from the start of the conversation.
	SecondsFromStart float64 `json:"secondsFromStart" url:"secondsFromStart"`
}

type BotMessage struct {
	// The role of the bot in the conversation.
	Role string `json:"role" url:"role"`
	// The message content from the bot.
	Message string `json:"message" url:"message"`
	// The timestamp when the message was sent.
	Time float64 `json:"time" url:"time"`
	// The timestamp when the message ended.
	EndTime float64 `json:"endTime" url:"endTime"`
	// The number of seconds from the start of the conversation.
	SecondsFromStart float64 `json:"secondsFromStart" url:"secondsFromStart"`
	// The source of the message.
	Source *string `json:"source,omitempty" url:"source,omitempty"`
	// The duration of the message in seconds.
	Duration *float64 `json:"duration,omitempty" url:"duration,omitempty"`
}
