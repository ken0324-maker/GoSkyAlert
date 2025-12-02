package services

import (
	"encoding/json"
	"final/models"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordService struct {
	Session    *discordgo.Session
	Amadeus    *AmadeusService
	Weather    *WeatherService
	Exchange   *ExchangeService
	Foursquare *FoursquareService
}

func NewDiscordService(token string, amadeus *AmadeusService, weather *WeatherService, exchange *ExchangeService, foursquare *FoursquareService) (*DiscordService, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	ds := &DiscordService{
		Session:    dg,
		Amadeus:    amadeus,
		Weather:    weather,
		Exchange:   exchange,
		Foursquare: foursquare,
	}

	dg.AddHandler(ds.handleMessage)
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsMessageContent

	return ds, nil
}

func (s *DiscordService) Start() error {
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

func formatTimeStr(ts string) string {
	if len(ts) >= 16 {
		return ts[11:16]
	}
	return ts
}

// 處理訊息
func (s *DiscordService) handleMessage(sess *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == sess.State.User.ID {
		return
	}

	args := strings.Fields(m.Content)
	if len(args) == 0 {
		return
	}

	command := args[0]

	switch command {
	case "!help", "/help":
		helpMsg := "**👋 GoSkyAlert 全能旅遊機器人**\n\n" +
			"✈️ **航班查詢**\n`/price [出發] [抵達] [日期]`\n範例：`/price TPE NRT 2026-03-01`\n\n" +
			"💱 **匯率查詢**\n`/rate [持有貨幣] [目標貨幣] (金額)`\n範例：`/rate USD TWD` 或 `/rate JPY TWD 1000`\n\n" +
			"🌤️ **天氣查詢**\n`/weather [城市名稱]`\n範例：`/weather Tokyo` 或 `/weather 台北`\n\n" +
			"🏛️ **景點搜尋**\n`/spot [城市/地點]`\n範例：`/spot 大阪` 或 `/spot 101大樓`"
		sess.ChannelMessageSend(m.ChannelID, helpMsg)

	// --- 航班查詢 ---
	case "!price", "/price":
		if len(args) < 4 {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 格式錯誤。\n請使用：`/price TPE NRT 2026-03-01`")
			return
		}
		origin := strings.ToUpper(args[1])
		dest := strings.ToUpper(args[2])
		date := args[3]

		sess.ChannelTyping(m.ChannelID)
		sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🔍 正在搜尋 **%s ➝ %s** (%s) 的航班...", origin, dest, date))

		req := models.SearchRequest{Origin: origin, Destination: dest, DepartureDate: date, Adults: 1, Currency: "TWD"}
		
		// [修正] 這裡接收 3 個回傳值：flights, advice, err
		flights, advice, err := s.Amadeus.SearchFlights(req)
		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ 搜尋失敗: %v", err))
			return
		}
		if len(flights) == 0 {
			sess.ChannelMessageSend(m.ChannelID, "📭 找不到航班。")
			return
		}

		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("✈️ **%s ➝ %s (%s)** 搜尋結果：\n", origin, dest, date))

		// [新增] 顯示價格建議
		if advice != nil {
			msg.WriteString(fmt.Sprintf("\n💡 **分析建議**: %s\n", advice.Advice))
		}

		limit := 3
		if len(flights) < limit {
			limit = len(flights)
		}
		for i := 0; i < limit; i++ {
			f := flights[i]
			msg.WriteString(fmt.Sprintf("\n**%d. %s (%s)**\n💰 **$%.0f %s** | ⏱️ %s\n%s %s ➝ %s %s\n",
				i+1, f.Airline, f.FlightNumber, f.Price, f.Currency, f.Duration,
				f.From.Code, formatTimeStr(f.Departure), f.To.Code, formatTimeStr(f.Arrival)))
		}
		sess.ChannelMessageSend(m.ChannelID, msg.String())

	// --- 匯率查詢 ---
	case "!rate", "/rate":
		if s.Exchange == nil {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 匯率服務未啟用")
			return
		}
		if len(args) < 3 {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 格式錯誤。\n請使用：`/rate USD TWD` 或 `/rate JPY TWD 1000`")
			return
		}
		from := strings.ToUpper(args[1])
		to := strings.ToUpper(args[2])
		amount := 1.0
		if len(args) >= 4 {
			if val, err := strconv.ParseFloat(args[3], 64); err == nil {
				amount = val
			}
		}

		sess.ChannelTyping(m.ChannelID)
		res, err := s.Exchange.GetExchangeRates(from, []string{to})
		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, "❌ 匯率查詢失敗")
			return
		}
		rate := res.Rates[to]
		converted := amount * rate

		msg := fmt.Sprintf("💱 **匯率換算**\n\n1 %s = %.4f %s\n\n💰 **%.2f %s ≈ %.2f %s**",
			from, rate, to, amount, from, converted, to)

		sess.ChannelMessageSend(m.ChannelID, msg)

	// --- 天氣查詢 ---
	case "!weather", "/weather":
		if s.Weather == nil {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 天氣服務未啟用")
			return
		}
		if len(args) < 2 {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 請輸入城市名稱，例如：`/weather Tokyo`")
			return
		}
		city := strings.Join(args[1:], " ")

		sess.ChannelTyping(m.ChannelID)
		wData, err := s.Weather.GetCurrentWeather(city)
		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, "❌ 找不到該城市天氣資訊")
			return
		}

		msg := fmt.Sprintf("🌤️ **%s (%s) 目前天氣**\n\n🌡️ 氣溫: **%.1f°C** (體感 %.1f°C)\n☁️ 狀況: %s\n💧 濕度: %d%%\n🌬️ 風速: %.1f km/h",
			wData.Location.Name, wData.Location.Country,
			wData.Current.TempC, wData.Current.FeelsLikeC,
			wData.Current.Condition.Text,
			wData.Current.Humidity,
			wData.Current.WindKph)
		sess.ChannelMessageSend(m.ChannelID, msg)

	// --- 景點搜尋 ---
	case "!spot", "/spot":
		if s.Foursquare == nil {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 景點服務未啟用")
			return
		}
		if len(args) < 2 {
			sess.ChannelMessageSend(m.ChannelID, "⚠️ 請輸入地點，例如：`/spot 東京`")
			return
		}
		locationName := strings.Join(args[1:], " ")

		sess.ChannelTyping(m.ChannelID)

		lat, lng, formattedName, err := getCoordinates(locationName)
		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ 找不到地點「%s」", locationName))
			return
		}

		// 這裡使用 services.SearchRequest
		spots, err := s.Foursquare.SearchNearby(SearchRequest{
			Latitude:  lat,
			Longitude: lng,
			Radius:    3000,
			Category:  "16000",
		})

		if err != nil {
			sess.ChannelMessageSend(m.ChannelID, "❌ 景點搜尋失敗")
			return
		}

		if len(spots) == 0 {
			sess.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📭 在 **%s** 附近沒找到景點。", formattedName))
			return
		}

		var msg strings.Builder
		msg.WriteString(fmt.Sprintf("🏛️ **%s** 附近的熱門景點：\n", formattedName))

		limit := 5
		if len(spots) < limit {
			limit = len(spots)
		}

		for i := 0; i < limit; i++ {
			spot := spots[i]
			dist := fmt.Sprintf("%.0fm", spot.Distance)
			if spot.Distance > 1000 {
				dist = fmt.Sprintf("%.1fkm", spot.Distance/1000)
			}
			msg.WriteString(fmt.Sprintf("\n**%d. %s**\n📍 距離: %s\n", i+1, spot.Name, dist))
		}
		sess.ChannelMessageSend(m.ChannelID, msg.String())
	}
}

// 輔助函式：使用 OpenStreetMap 進行簡易 Geocoding
func getCoordinates(query string) (float64, float64, string, error) {
	url := fmt.Sprintf("https://nominatim.openstreetmap.org/search?format=json&q=%s&limit=1", url.QueryEscape(query))
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	var results []struct {
		Lat         string `json:"lat"`
		Lon         string `json:"lon"`
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, "", err
	}
	if len(results) == 0 {
		return 0, 0, "", fmt.Errorf("not found")
	}

	lat, _ := strconv.ParseFloat(results[0].Lat, 64)
	lon, _ := strconv.ParseFloat(results[0].Lon, 64)

	displayName := strings.Split(results[0].DisplayName, ",")[0]

	return lat, lon, displayName, nil
}
