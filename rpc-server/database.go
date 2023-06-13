package main

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connect() {
	dsn := "user:password@tcp(mysql_db:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
	connection, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB = connection
	DB.AutoMigrate(&Message{})
}

func sendMessage(chatId string, content string, sender string) error {
	message := Message{
		ChatId:   chatId,
		Content:  content,
		Sender:   sender,
		SendTime: time.Now().Unix(),
	}
	if err := DB.Create(&message).Error; err != nil {
		return err
	}
	return nil
}

func pullMessage(chatId string, limit int32, cursor int64, reverse bool) ([]*Message, error) {
	var messages []*Message
	var order string
	if reverse {
		order = "desc"
	} else {
		order = "asc"
	}
	err := DB.Where("chat_id", chatId).Order(fmt.Sprintf("send_time %s", order)).Limit(int(limit) + 1).Offset(int(cursor)).Find(&messages).Error
	if err != nil {
		return nil, err
	}
	return messages, nil
}
