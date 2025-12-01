package services

import (
	"final/models"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordService struct {
	Session *discordgo.Session
	Amadeus *AmadeusService
}

func NewDiscordService(token string, amadeus *AmadeusService) (*DiscordService, error) {
	// 建立 Discord Session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	ds := &DiscordService{
		Session: dg,
		Amadeus: amadeus,
	}

	// 註冊訊息處理函式
	dg.AddHandler(ds.handleMessage)

	// 設定 Intent (必須包含 MessageContent 才能讀取訊息內容)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	return ds, nil
}

func (s *DiscordService) Start() error {
	// 開啟 WebSocket 連線
	err := s.Session.Open()
	if err != nil {
		return fmt.Errorf("開啟 Discord 連線失敗: %v", err)
	}
	log.Println("🤖 Discord Bot 已連線！")
	return nil
}

func (s *DiscordService) Stop() {
	s.Session.Close()
}

// [新增] 輔助函式：從 "2026-02-14T12:10:00" 提取 "12:10"
func formatTimeStr(ts string) string {
	// 確保字串長度足夠，避免 panic
	if len(ts) >= 16 {
		// 取出 T 之後的時間部分 HH:MM
		return ts[11:16]
	}
	return ts
}

// 處理訊息
func (s *DiscordService) handleMessage(sess *discordgo.Session, m *discordgo.MessageCreate) {
	// 忽略機器人自己發送的訊息
	if m.Author.ID == sess.State.User.ID {
		return
	}

	// 簡單的指令解析
	args := strings.Fields(m.Content)
	if len(args) == 0 {
		return
	}

	command := args[0]

	switch command {
	case "!help", "/help":
		helpMsg := "**👋 歡迎使用 GoSkyAlert 航班機器人！**\n\n" +
			"請輸入以下指令查詢：\n" +
			"`/price [出發地] [目的地] [日期]`\n" +
			"範例：`/price TPE NRT 2026-03-01`"
		sess.ChannelMessageSend(m.ChannelID, helpMsg)

	case "!price", "/price":
		if len(args) < 4 {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 格式錯誤。\n請使用：`/price TPE NRT 2026-03-01`")
			return
		}

		origin := strings.ToUpper(args[1])
		dest := strings.ToUpper(args[2])
		date := args[3]

		// 發送 "正在輸入..." 狀態
		sess.ChannelTyping(m.ChannelID)
		sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔍 正在搜尋 **%s ➝ %s** (%s) 的航班...", origin, dest, date))

		// 呼叫 Amadeus 搜尋
		req := models.SearchRequest{
			Origin:        origin,
			Destination:   dest,
			DepartureDate: date,
			Adults:        1,
			Currency:      "TWD",
		}

		flights, err := s.Amadeus.SearchFlights(req)
		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ 搜尋失敗: %v", err))
			return
		}

		if len(flights) == 0 {
			sess.ChannelMessageSend(m.ChannelID, "📭 找不到符合條件的航班，請嘗試其他日期。")
			return
		}

		// 構建回應訊息
		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("✈️ **%s ➝ %s (%s)** 搜尋結果：\n\n", origin, dest, date))

		limit := 3
		if len(flights) < limit {
			limit = len(flights)
		}

		for i := 0; i < limit; i++ {
			f := flights[i]
			msg.WriteString(fmt.Sprintf("**%d. %s (%s)**\n", i+1, f.Airline, f.FlightNumber))
			msg.WriteString(fmt.Sprintf("💰 價格: **$%.0f %s**\n", f.Price, f.Currency))

			// [修改] 使用 formatTimeStr 處理字串時間，正確顯示 HH:MM
			depTime := formatTimeStr(f.Departure)
			arrTime := formatTimeStr(f.Arrival)

			msg.WriteString(fmt.Sprintf("⏱️ 時間: %s ➝ %s (%s)\n", depTime, arrTime, f.Duration))
			msg.WriteString("------------------------------\n")
		}

		msg.WriteString(fmt.Sprintf("\n📊 共找到 %d 個航班。", len(flights)))

		sess.ChannelMessageSend(m.ChannelID, msg.String())
	}
}
