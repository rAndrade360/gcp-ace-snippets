package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

var (
	PORT       = os.Getenv("PORT")
	PROJECT_ID = os.Getenv("PROJECT_ID")
)

func main() {
	if PORT == "" {
		PORT = "8080"
		log.Printf("Defaulting to PORT %s", PORT)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	client, err := datastore.NewClient(context.Background(), PROJECT_ID)
	if err != nil {
		log.Fatal(err)
	}

	dtclient := NewDatastore(client, "Message")

	bot.Debug = true

	updates := bot.ListenForWebhook("/bot")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	log.Printf("Authorized on account %s", bot.Self.UserName)

	go http.ListenAndServe(":"+PORT, nil)

	receive := func(b []byte) error {
		var rec MessageReceived

		err = json.Unmarshal(b, &rec)
		if err != nil {
			return err
		}

		fmt.Println("Received data: ", rec)

		if rec.Type != "GENERATED_ASCII" {
			return errors.New("forced error")
		}

		msg, err := dtclient.GetMessageByID(context.Background(), rec.ID)
		if err != nil {
			log.Println("Nao peguei a msg: ", err.Error())
			return err
		}
		
		log.Println("MSG: ", msg)

		msgTxt := tgbotapi.NewMessage(msg.ChatID, rec.RawImage)
		msgTxt.ReplyToMessageID = msg.MessageID

		bot.Send(msgTxt)
		return nil
	}

	go Sub(context.Background(), PROJECT_ID, receive)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if len(update.Message.Photo) > 0 {

				file, err := bot.GetFile(tgbotapi.FileConfig{
					FileID: update.Message.Photo[0].FileID,
				})
				if err != nil {
					txt := "Não foi dessa vez, estamos com problemas internos"
					msgTxt := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
					msgTxt.ReplyToMessageID = update.Message.MessageID
					bot.Send(msgTxt)
					continue
				}

				mes := Message{
					ID:        uuid.NewString(),
					Type:      "UPLOAD_FROM_TELEGRAM",
					FilePath:  file.FilePath,
					ChatID:    update.Message.Chat.ID,
					MessageID: update.Message.MessageID,
				}

				err = dtclient.SaveMessage(context.Background(), mes)
				if err != nil {
					log.Println("Err to save msg: ", err.Error())
					txt := "Não foi dessa vez, estamos com problemas internos"
					msgTxt := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
					msgTxt.ReplyToMessageID = update.Message.MessageID

					bot.Send(msgTxt)
				}

				d, _ := json.Marshal(mes)

				err = Pub(context.Background(), PROJECT_ID, d)
				txt := "Aguarde um pouco, quando finalizarmos o processamento te enviaremos o resultado."
				if err != nil {
					txt = "Não foi dessa vez, estamos com problemas internos"
				}

				msgTxt := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
				msgTxt.ReplyToMessageID = update.Message.MessageID

				bot.Send(msgTxt)
				continue
			}

			txt := fmt.Sprintf("Olá %s, que tal enviar uma imagem para ver um truque de mágica?", update.Message.From.FirstName)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
