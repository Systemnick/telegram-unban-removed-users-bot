package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
)

func main() {
	settings := tele.Settings{
		Token:  os.Getenv("TELEGRAM_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(settings)
	if err != nil {
		log.Fatal(err)
		return
	}

	i, _ := strconv.Atoi(os.Getenv("ADMIN_GROUP_ID"))
	adminGroupID := int64(i)

	b.Handle(tele.OnUserLeft, func(c tele.Context) error {
		msg := c.Message()
		sender := msg.Sender
		leftUser := msg.UserLeft

		if sender == nil || leftUser == nil || sender.ID == leftUser.ID {
			return nil
		}

		err = b.Unban(c.Chat(), leftUser)
		if err != nil {
			return err
		}

		who := getUserDescriptionMD(sender)
		whom := getUserLinkMD(leftUser)
		chat := getGroupLinkMD(c.Message().Chat)

		log.Printf("Sender: %+v", *sender)
		log.Printf("Sender MD: %s", who)
		log.Printf("UserLeft: %+v", *leftUser)
		log.Printf("UserLeftMD: %s", whom)

		if sender.ID == leftUser.ID {
			text := fmt.Sprintf("%s left %s", whom, chat)

			_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

			return err
		}

		text := fmt.Sprintf("Unban %s in %s removed by %s", whom, chat, who)

		_, err = b.Send(&tele.Chat{ID: adminGroupID}, text, tele.ModeMarkdownV2)

		return err
	})

	b.Start()
}

func getUserTitle(user *tele.User) string {
	return strings.Trim(fmt.Sprintf("%s %s", user.FirstName, user.LastName), " ")
}

func getUserDescriptionMD(user *tele.User) string {
	text := getUserTitle(user)
	return escapeSpecialCharactersMD(fmt.Sprintf(`%s (%s)`, text, user.Username))
}

func getUserLinkMD(user *tele.User) string {
	text := getUserTitle(user)
	return fmt.Sprintf("[%s](tg://user?id=%d)", escapeSpecialCharactersMD(text), user.ID)
}

func getGroupLinkMD(c *tele.Chat) string {
	text := escapeSpecialCharactersMD(c.Title)
	if c.Username == "" {
		return text
	}

	return fmt.Sprintf("[%s](tg://resolve?domain=%s)", text, c.Username)
}

func getPinnedMessageLinkMD(m *tele.Message) string {
	return fmt.Sprintf("[message](tg://resolve?domain=%s&post=%d&single)", m.ReplyTo.Chat.Username, m.ReplyTo.ID)
}

// In all other places characters:
// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
// must be escaped with the preceding character '\'.
var specialCharactersMD = "_*[]()~`>#+-=|{}.!"

func escapeSpecialCharactersMD(text string) string {
	prevIdx := 0

	for {
		idx := strings.IndexAny(text[prevIdx:], specialCharactersMD)
		if idx == -1 {
			break
		}

		idx += prevIdx
		text = text[:idx] + `\` + text[idx:idx] + text[idx:]
		prevIdx = idx + 2
	}

	return text
}
