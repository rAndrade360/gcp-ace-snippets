package main

type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	FilePath  string `json:"filePath"`
	ChatID    int64  `json:"chatId"`
	MessageID int    `json:"messageId"`
}

type MessageReceived struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	RawImage string `json:"rawImage"`
}
