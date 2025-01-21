package main

import (
	"log"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	s "github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var lastLink string
var start bool = false
var channelID int64 //id вашего канала
var imParseNow bool = false

func main() {
	bot, err := tg.NewBotAPI("ваш токен")
	if err != nil {
		log.Println("Ошибка токена: ", err)
		return
	}

	u := tg.NewUpdate(60)
	u.Timeout = 0
	updates := bot.GetUpdatesChan(u)

	keyboard := tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("старт"),
		),
	)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		switch update.Message.Text {
		case "старт":
			if !start {
				start = true
			
				go itsTimeToStart(update, bot)

				keyboard = tg.NewReplyKeyboard(
					tg.NewKeyboardButtonRow(
						tg.NewKeyboardButton("стоп"),
					),
				)
			}
			
		case "стоп":
			if !imParseNow {
				start = false
				
				keyboard = tg.NewReplyKeyboard(
					tg.NewKeyboardButtonRow(
						tg.NewKeyboardButton("старт"),
					),
				)
				sendMessageBot("парсер остановлен", update, bot, keyboard)
			}else {
				sendMessageBot("парсер пока остановить нельзя, подождите пару минут", update, bot, keyboard)
			}
		}
		sendMessageBot("ждём", update, bot, keyboard)
	}
}

func itsTimeToStart( update tg.Update, bot *tg.BotAPI) {
	for start {
		link := parseHabr()
			if link == lastLink {
				log.Println("ничего нового(")
			} else {
				sendMessageChannel("что то новое)", update, bot)
				sendMessageChannel(link, update, bot)
				lastLink = link
				msg := tg.NewMessage(update.Message.Chat.ID, "ждём")
				bot.Send(msg)
				msg.ReplyMarkup = tg.NewReplyKeyboard(
					tg.NewKeyboardButtonRow(
						tg.NewKeyboardButton("стоп"),
					),
				)				
			}
					
			time.Sleep(1 * time.Hour)
	}
}

func parseHabr() string{
	imParseNow = true
	service, err := s.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		log.Println("ошибка с сервисом: ", err)
		return "ничего"
	}
	defer service.Stop()

	caps := s.Capabilities{"browserName": "chrome"}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"window-size=1920x1080",
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-blink-features=AutomationControlled",
			"--disable-infobars",
		},
	})

	driver, err := s.NewRemote(caps, "")
	if err != nil {
		log.Println("проблемы с драйвером: ", err)
		return "ничего"
	}

	err = driver.Get("https://habr.com/ru/search/?q=go&target_type=posts&order=date")
	if err != nil {
		log.Println("Не удалось открыть страницу")
		return "Не удалось открыть страницу"
	}
	defer driver.Close()

	article, err := driver.FindElement(s.ByCSSSelector, "article")
	if err != nil {
		log.Println("не смог найти артикуль: ", err)
		return "не смог найти ссылку"
	}

	h2, err := article.FindElement(s.ByCSSSelector, "h2")
	if err != nil {
		log.Println("не смог найти h2: ", err)
		return "не смог найти ссылку"
	}
	
	href, err := h2.FindElement(s.ByCSSSelector, "a")
	if err != nil {
		log.Println("не смог найти a: ", err)
		return "не смог найти ссылку"
	}

	link, _ := href.GetAttribute("href")

	log.Println(link)
	imParseNow = false
	return link
}

func sendMessageChannel(text string, update tg.Update, bot *tg.BotAPI) {
	msgChannel := tg.NewMessage(channelID, text)
	bot.Send(msgChannel)

}

func sendMessageBot(text string, update tg.Update, bot *tg.BotAPI, keyboard tg.ReplyKeyboardMarkup) {
	msg := tg.NewMessage(update.Message.Chat.ID, text)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)	
}