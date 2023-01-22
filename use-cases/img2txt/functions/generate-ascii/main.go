package generateascii

import (
	"context"
	"encoding/json"
	"image"
	"image/jpeg"
	"math"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/nfnt/resize"
)

type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type Message struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	RawImage string `json:"rawImage"`
}

var PROJECT_ID = os.Getenv("PROJECT_ID")

var scale = []string{" ", ".", ",", "-", "~", "+", "=", "7", "8", "9", "$", "W", "#", "@", "Ñ"}

func Generate(ctx context.Context, event GCSEvent) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	defer client.Close()

	o := client.Bucket(event.Bucket).Object(event.Name)
	attrs, err := o.Attrs(ctx)
	if err != nil {
		return err
	}
	messageId, ok := attrs.Metadata["messageId"]
	if !ok {
		return nil
	}

	r, err := o.NewReader(ctx)
	if err != nil {
		return err
	}

	defer r.Close()

	var raw_img string
	if attrs.ContentType == "image/jpeg" {
		img, err := jpeg.Decode(r)
		if err != nil {
			return err
		}

		raw_img = GenerateASCII(img)
	}

	if len(raw_img) > 0 {
		msg := Message{
			ID:       messageId,
			RawImage: raw_img,
			Type:     "GENERATED_ASCII",
		}

		d, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		err = Pub(ctx, PROJECT_ID, d)
		if err != nil {
			return err
		}
	}

	return nil
}

func GenerateASCII(imge image.Image) string {
	imge = resize.Resize(85, 40, imge, resize.Lanczos2)

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
