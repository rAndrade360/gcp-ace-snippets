package main

type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	FileID    string `json:"fileId"`
	BotToken  string `json:"botToken"`
	ChatID    int64  `json:"chatId"`
	MessageID int    `json:"messageId"`
}
