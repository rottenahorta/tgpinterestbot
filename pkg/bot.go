package bot

import (
	"io"
	"log"
	"math/rand"
	"net/http"

	gj "github.com/tidwall/gjson"
	sl "golang.org/x/exp/slices"
	tb "gopkg.in/telebot.v3"
)

type Bot struct {
	bot  *tb.Bot
	stop chan chan struct{}
}

var (
	authIDs = []int64{450892706, 414980848, 862465186, 694585332}
	pinURLs = []string{"https://api.pinterest.com/v3/pidgets/boards/marygapon/places/pins/", 
					"https://api.pinterest.com/v3/pidgets/boards/marygapon/home/pins/", 
					"https://api.pinterest.com/v3/pidgets/boards/prijmakkarina2015/интерьер/pins",
					} 
)

func NewBot(b *tb.Bot) *Bot {
	return &Bot{bot: b}
}

func (b *Bot) Start() {
	b.handleMsg()
	if b.bot.Poller == nil {
		panic("telebot: can't start without a poller")
	}
	stop := make(chan struct{})
	stopConfirm := make(chan struct{})
	go func() {
		b.bot.Poller.Poll(b.bot, b.bot.Updates, stop)
		close(stopConfirm)
	}()
	for {
		select {
		case upd := <-b.bot.Updates:
			b.bot.ProcessUpdate(upd)
		case confirm := <-b.stop:
			close(stop)
			<-stopConfirm
			close(confirm)
			return
		}
	}
}

func (b *Bot) handleMsg() {
	b.bot.Handle(tb.OnText, func(c tb.Context) error {
		if !sl.Contains(authIDs, c.Sender().ID) {
			log.Printf("Unauthorized account %d", c.Sender().ID)
			return c.Send("u r not in dbv!")
		}
		log.Printf("Authorized account %d %s", c.Sender().ID, c.Message().Text)
		c.Delete()

		var urlsl []gj.Result
		for _, u := range pinURLs {
			r, err := http.Get(u) //https://api.pinterest.com/v5/boards/473441048266790946/pins/
			if err != nil {
				log.Printf("error dumpin pinterest board %s", err)
			}
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("error readin pinterest body res %s", err)
			}

			urlsl = append(urlsl, gj.Get(string(body), "data.pins.#.images.564x.url").Array()...)
		}
		ind := rand.Intn(50*len(pinURLs))
		log.Printf("photo #%d", ind)
		return c.Send(&tb.Photo{File: tb.FromURL(urlsl[ind].String()), Caption: "#dbv"})
	})
}
