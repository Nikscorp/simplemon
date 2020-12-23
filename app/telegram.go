package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

const DefaultHTTPTimeoutSec = 5

type Telegram struct {
	Bot                    *tb.Bot
	Recipients             []string
	SuccessConfidence      int
	FailConfidence         int
	LastTaskNotifiedResult map[string]bool
}

func NewTelegram(conf TelegramConf) Telegram {
	t := Telegram{}
	bot, err := tb.NewBot(tb.Settings{
		Token:  conf.Token,
		Client: &http.Client{Timeout: DefaultHTTPTimeoutSec * time.Second},
	})
	if err != nil {
		log.Fatalf("[CRITICAL] Failed to init telegram bot api")
	}
	t.Bot = bot
	t.Recipients = conf.Recipients
	if conf.SuccessConfidence == 0 {
		t.SuccessConfidence = 1
	} else {
		t.SuccessConfidence = conf.SuccessConfidence
	}
	if conf.FailConfidence == 0 {
		conf.FailConfidence = 1
	} else {
		t.FailConfidence = conf.FailConfidence
	}
	t.LastTaskNotifiedResult = make(map[string]bool)
	return t
}

func (t Telegram) NotifyOrIgnore(ctx context.Context, tr TaskResult, ts map[string]TaskState) {
	if !tr.Task.Notify.Telegram {
		return
	}
	actualState := ts[tr.Task.ID]

	if !actualState.everChanged {
		return
	}
	maxSC := Max(tr.Task.SuccessConfidence, t.SuccessConfidence)
	maxFC := Max(tr.Task.FailConfidence, t.FailConfidence)

	lastNotifiedResult, isPresent := t.LastTaskNotifiedResult[tr.Task.ID]

	if !tr.WasError && actualState.sameResultCnt == maxSC && (lastNotifiedResult != tr.WasError && isPresent) {
		t.Notify(ctx, tr, maxSC)
		t.LastTaskNotifiedResult[tr.Task.ID] = tr.WasError
		return
	}

	if tr.WasError && actualState.sameResultCnt == maxFC && (lastNotifiedResult != tr.WasError || !isPresent) {
		t.Notify(ctx, tr, maxFC)
		t.LastTaskNotifiedResult[tr.Task.ID] = tr.WasError
		return
	}
	log.Printf("[INFO] Skip change state to %v telegram notifying for %s", tr.WasError, tr.Task.ID)
}

func (t *Telegram) Notify(ctx context.Context, tr TaskResult, confidenceLevel int) { //nolint: unparam
	log.Printf("[INFO] Notifying that task State changed to %v, with CL: %d", tr.WasError, confidenceLevel)
	text := t.GetMessageTextByTask(tr, confidenceLevel)
	file := &tb.Document{
		File:     tb.FromReader(strings.NewReader(tr.Output)),
		FileName: "log.txt",
	}
	for _, r := range t.Recipients {
		msg, err := t.Bot.Send(
			recipient{chatID: r},
			text,
		)
		if err != nil {
			log.Printf("[ERROR] Failed to send message, %v", err)
			return
		}
		_, err = file.Send(
			t.Bot, recipient{chatID: r},
			&tb.SendOptions{ReplyTo: msg},
		)
		if err != nil {
			log.Printf("[ERROR] Failed to send file, %v", err)
		}
	}
}

func (t *Telegram) GetMessageTextByTask(tr TaskResult, confidenceLevel int) string {
	msgTemplate := "[%s] %s\nDescription: %s\nConfidenceLevel: %d\n"
	state := "OK"
	if tr.WasError {
		state = "FAIL"
	}
	res := fmt.Sprintf(msgTemplate, state, tr.Task.ID, tr.Task.Description, confidenceLevel)
	return res
}

type recipient struct {
	chatID string
}

func (r recipient) Recipient() string {
	return r.chatID
}
