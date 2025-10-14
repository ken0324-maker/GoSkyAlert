package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type TelegramService struct {
	botToken string
}

func NewTelegramService(botToken string) *TelegramService {
	if botToken == "" {
		return nil // 如果沒有 token，返回 nil
	}
	return &TelegramService{
		botToken: botToken,
	}
}

// 發送航班通知 - 超級簡單版本
func (t *TelegramService) SendFlightNotification(chatID string, flights []map[string]interface{}) error {
	if t == nil {
		return nil // 如果服務未初始化，靜默返回
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	// 建立簡單的訊息
	message := "✈️ 找到航班啦！\n\n"

	if len(flights) > 0 {
		flight := flights[0] // 只顯示最便宜的一個
		message += fmt.Sprintf(
			"🛫 %s %s\n"+
				"📍 %s → %s\n"+
				"💰 %s %.0f %s\n"+
				"⏰ %s",
			flight["airline"],
			flight["flight_number"],
			flight["from"].(map[string]interface{})["code"],
			flight["to"].(map[string]interface{})["code"],
			"價格:",
			flight["price"],
			flight["currency"],
			"立即查看詳情！",
		)
	} else {
		message = "❌ 沒有找到符合條件的航班"
	}

	// 發送請求
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("text", message)

	_, err := http.PostForm(apiURL, data)
	return err
}

// 獲取 Chat ID 的簡單方法
func (t *TelegramService) GetChatID() (string, error) {
	if t == nil {
		return "", fmt.Errorf("Telegram service not initialized")
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", t.botToken)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// 簡單解析第一個訊息的 chat ID
	if updates, ok := result["result"].([]interface{}); ok && len(updates) > 0 {
		if firstUpdate, ok := updates[0].(map[string]interface{}); ok {
			if message, ok := firstUpdate["message"].(map[string]interface{}); ok {
				if chat, ok := message["chat"].(map[string]interface{}); ok {
					if chatID, ok := chat["id"].(float64); ok {
						return fmt.Sprintf("%.0f", chatID), nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("請先傳送訊息給您的 Bot")
}
