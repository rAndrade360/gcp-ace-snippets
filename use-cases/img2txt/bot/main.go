package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"math"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nfnt/resize"
)

var scale = []string{" ", ".", ",", "-", "~", "+", "=", "7", "8", "9", "$", "W", "#", "@", "Ñ"}

func main() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
		log.Printf("Defaulting to PORT %s", PORT)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	updates := bot.ListenForWebhook("/bot")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	log.Printf("Authorized on account %s", bot.Self.UserName)

	go http.ListenAndServe(":"+PORT, nil)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if len(update.Message.Photo) > 0 {
				msg := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileID(update.Message.Photo[0].FileID))

				msg.ReplyToMessageID = update.Message.MessageID

				f, err := bot.GetFile(tgbotapi.FileConfig{
					FileID: update.Message.Photo[0].FileID,
				})
				if err != nil {
					log.Println("Err to get file: ", err.Error())
					bot.Send(msg)
					continue
				}

				res, err := http.Get(f.Link(os.Getenv("BOT_TOKEN")))
				if err != nil {
					log.Println("Err to get file: ", err.Error())
					bot.Send(msg)
					continue
				}

				imge, err := jpeg.Decode(res.Body)
				if err != nil {
					log.Fatal("Err to decode img: ", err.Error())
					bot.Send(msg)
					continue
				}

				txt := GenerateASCII(imge)

				msgTxt := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
				msgTxt.ReplyToMessageID = update.Message.MessageID

				bot.Send(msgTxt)
				continue
			}

			txt := fmt.Sprintf("Olá %s, que tal enviar um imagem para ver um truque de mágica?", update.Message.From.FirstName)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, txt)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}

func GenerateASCII(imge image.Image) string {
	imge = resize.Resize(38, 17, imge, resize.Lanczos2)

	txt := ""

	for y := imge.Bounds().Min.Y; y <= imge.Bounds().Max.Y; y++ {
		for x := imge.Bounds().Min.X; x <= imge.Bounds().Max.X; x++ {
			c := imge.At(x, y)
			r, g, b, _ := c.RGBA()

			gray := (r + g + b) / 3

			char := int(math.Round((float64(gray) / 65536) * float64(len(scale)-1)))
			txt += string(scale[char])
		}
		txt += "\n"
	}

	return txt
}
