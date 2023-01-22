package storagedeploy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

type Message struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	FilePath  string `json:"filePath"`
	ChatID    int64  `json:"chatId"`
	MessageID int    `json:"messageId"`
}

var (
	bucket     = os.Getenv("BUCKET_NAME")
	objectPath = os.Getenv("OBJECT_PATH")
)

func Deploy(ctx context.Context, msg PubSubMessage) error {
	var m Message

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	err := json.Unmarshal(msg.Data, &m)
	if err != nil {
		return err
	}

	if m.Type != "UPLOAD_FROM_TELEGRAM" {
		return nil
	}

	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("BOT_TOKEN"), m.FilePath)
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	var filename = m.FilePath
	ss := strings.Split(m.FilePath, "/")
	if len(ss) > 0 {
		filename = m.ID + "-" + ss[len(ss)-1]
	}

	w := client.Bucket(bucket).Object(objectPath + filename).NewWriter(ctx)
	defer w.Close()
	_, err = io.Copy(w, res.Body)
	if err != nil {
		return err
	}

	w.Close()

	return nil
}
