package main

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

var mClient *mixin.Client

func init() {
	k := config.Mixin
	var err error
	mClient, err = mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   k.ClientID,
		SessionID:  k.SessionID,
		PrivateKey: k.PrivateKey,
	})
	if err != nil {
		panic(err)
	}
}
func SendMsgToDeveloper(ctx context.Context, msg string) {
	userID := config.Dev
	if userID == "" {
		return
	}

	conversationID := mixin.UniqueConversationID(config.Mixin.ClientID, userID)
	_ = mClient.SendMessage(ctx, &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           Base64Encode([]byte(msg)),
	})
}

func GetUUID() string {
	id, _ := uuid.NewV4()
	return id.String()
}

func Base64Encode(str []byte) string {
	return base64.RawURLEncoding.EncodeToString(str)
}

func GetZeroTimeByDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func GetMonthStartByDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
