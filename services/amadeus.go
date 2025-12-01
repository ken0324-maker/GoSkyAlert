package services

import (
	"encoding/json"
	"final/config"
	"final/models"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os" // 引入 os 模組用於檔案操作
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 設定歷史記錄檔案路徑
const historyFilePath = "amadeus_api_history.jsonl"

type AmadeusService struct {
	config        *config.Config
	client        *http.Client
	accessToken   string
	tokenExpiry   time.Time
	trackingData  map[string]*models.PriceAnalysis
	trackingMutex sync.RWMutex
}

func NewAmadeusService(cfg *config.Config) *AmadeusService {
	return &AmadeusService{
		config:       cfg,
		client:       &http.Client{Timeout: 30 * time.Second},
		trackingData: make(map[string]*models.PriceAnalysis),
	}
}

// 新增：將 API 響應儲存到本地歷史記錄檔案
// 採用 JSON Lines (.jsonl) 格式，每次寫入一行 JSON
func (s *AmadeusService) saveApiHistory(origin, destination, departureDate string, rawBody []byte) {
	// 構造要儲存的歷史記錄結構
	historyEntry := struct {
		Timestamp     time.Time       `json:"timestamp"`
		Origin        string          `json:"origin"`
		Destination   string          `json:"destination"`
		DepartureDate string          `json:"departure_date"`
		RawResponse   json.RawMessage `json:"raw_response"`
	}{
		Timestamp:     time.Now(),
		Origin:        origin,
		Destination:   destination,
		DepartureDate: departureDate,
		RawResponse:   rawBody,
	}

	// 序列化為 JSON
	jsonLine, err := json.Marshal(historyEntry)
	if err != nil {
		log.Printf("❌ 歷史記錄序列化失敗: %v", err)
		return
	}

	// 開啟檔案，如果不存在則創建，O_APPEND 模式用於追加
	file, err := os.OpenFile(historyFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ 無法開啟歷史記錄檔案 %s: %v", historyFilePath, err)
		return
	}
	defer file.Close()

	// 寫入 JSON 行和換行符
	_, err = file.Write(jsonLine)
	if err != nil {
		log.Printf("❌ 寫入歷史記錄失敗: %v", err)
		return
	}
	_, err = file.WriteString("\n")
	if err != nil {
		log.Printf("❌ 寫入換行符失敗: %v", err)
		return
	}

	log.Printf("💾 成功將 API 響應儲存到歷史記錄檔案: %s", historyFilePath)
}

// 新增：機票價格追蹤功能
// 修改：使用真實 API 進行價格追蹤
func (s *AmadeusService) TrackFlightPrices(req models.PriceTrackingRequest) (*models.PriceAnalysis, error) {
	route := fmt.Sprintf("%s-%s", req.Origin, req.Destination)

	log.Printf("🔄 開始真實價格追蹤: %s, 週數: %d", route, req.Weeks)

	analysis := &models.PriceAnalysis{
		Route:      route,
		TrackWeeks: req.Weeks,
		CreatedAt:  time.Now(),
	}

	// 使用真實 API 查詢每週價格
	for week := 1; week <= req.Weeks; week++ {
		searchDate := time.Now().AddDate(0, 0, (week-1)*7)
		travelDate := searchDate.AddDate(0, 0, 30) // 假設30天後出發

		log.Printf("🔍 查詢第 %d 週 - 搜索日期: %s, 出發日期: %s",
			week, searchDate.Format("2006-01-02"), travelDate.Format("2006-01-02"))

		// 使用真實 API 獲取價格
		price, err := s.getRealTimePrice(req.Origin, req.Destination, travelDate.Format("2006-01-02"))
		if err != nil {
			log.Printf("⚠️ 第 %d 週 API 查詢失敗: %v", week, err)
			// 如果 API 失敗，使用智能估算
			price = s.estimatePrice(req.Origin, req.Destination, travelDate, week, req.Weeks)
		}

		dataPoint := models.PricePoint{
			Week:     week,
			Date:     travelDate,
			Price:    price,
			Currency: "TWD",
		}

		analysis.DataPoints = append(analysis.DataPoints, dataPoint)

		log.Printf("💰 第 %d 週 - 出發: %s, 價格: $%.0f",
			week, travelDate.Format("2006-01-02"), price)

		// 如果是真實長期追蹤，這裡會暫停
		// time.Sleep(1 * time.Second) // 避免 API 限制
	}

	// 計算統計數據
	s.calculatePriceStatistics(analysis)

	log.Printf("✅ 真實價格追蹤完成: %s, 最低價: $%.0f", route, analysis.MinPrice)
	return analysis, nil
}

// 新增：實時價格查詢
// 修改：過濾重複航空公司，只取每個航空公司的最低價格
// 完整的 getRealTimePrice 方法
func (s *AmadeusService) getRealTimePrice(origin, destination, departureDate string) (float64, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return 0, err
	}

	apiURL := fmt.Sprintf("%s/shopping/flight-offers", s.config.AmadeusBaseURL)

	params := url.Values{}
	params.Add("originLocationCode", origin)
	params.Add("destinationLocationCode", destination)
	params.Add("departureDate", departureDate)
	params.Add("adults", "1")
	params.Add("currencyCode", "TWD")
	params.Add("max", "20") // 增加數量以獲得更多選擇

	fullURL := apiURL + "?" + params.Encode()

	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("創建請求失敗: %v", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	log.Printf("📡 呼叫真實 API: %s -> %s, 日期: %s", origin, destination, departureDate)

	// 發送請求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("讀取響應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ API錯誤: 狀態碼 %d, 響應: %s", resp.StatusCode, string(body))
		return 0, fmt.Errorf("API錯誤: 狀態碼 %d", resp.StatusCode)
	}

	// !!! 新增功能：將 API 響應儲存到歷史記錄檔案
	s.saveApiHistory(origin, destination, departureDate, body)
	// !!!

	// 解析響應
	var apiResponse models.AmadeusFlightOffersResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return 0, fmt.Errorf("解析JSON失敗: %v", err)
	}

	if len(apiResponse.Data) == 0 {
		return 0, fmt.Errorf("未找到航班")
	}

	// 使用 map 來儲存每個航空公司的唯一最低價格
	airlinePrices := make(map[string]float64)
	var uniqueAirlines []string

	log.Printf("   📊 原始找到 %d 個航班", len(apiResponse.Data))

	for i, offer := range apiResponse.Data {
		if len(offer.Itineraries) == 0 || len(offer.Itineraries[0].Segments) == 0 {
			continue
		}

		price, err := strconv.ParseFloat(offer.Price.Total, 64)
		if err != nil {
			continue
		}

		carrierCode := offer.Itineraries[0].Segments[0].CarrierCode
		airline := s.getAirlineName(carrierCode)

		// 如果這個航空公司還沒有記錄，或者找到更低的價格，就更新
		if existingPrice, exists := airlinePrices[airline]; !exists || price < existingPrice {
			airlinePrices[airline] = price
		}

		// 記錄獨特航空公司
		if !contains(uniqueAirlines, airline) {
			uniqueAirlines = append(uniqueAirlines, airline)
		}

		log.Printf("   ✈️ 航班 %d: %s (%s) - $%.0f", i+1, airline, carrierCode, price)
	}

	// 轉換為價格列表來計算平均
	var prices []float64
	var airlines []string

	log.Printf("   🔄 過濾後得到 %d 個獨特航空公司:", len(airlinePrices))
	for airline, price := range airlinePrices {
		prices = append(prices, price)
		airlines = append(airlines, airline)
		log.Printf("      ✅ %s: $%.0f", airline, price)
	}

	if len(prices) == 0 {
		return 0, fmt.Errorf("無法解析航班價格")
	}

	// 使用所有過濾後的獨特航空公司計算平均價格
	totalPrice := 0.0
	validAirlines := len(airlinePrices)

	log.Printf("   📈 使用 %d 個獨特航空公司計算平均價格:", validAirlines)

	// 按價格排序以便顯示
	sortedAirlines := make([]struct {
		airline string
		price   float64
	}, 0, len(airlinePrices))

	for airline, price := range airlinePrices {
		sortedAirlines = append(sortedAirlines, struct {
			airline string
			price   float64
		}{airline, price})
		totalPrice += price
	}

	// 按價格排序
	sort.Slice(sortedAirlines, func(i, j int) bool {
		return sortedAirlines[i].price < sortedAirlines[j].price
	})

	// 顯示排序後的價格
	for i, item := range sortedAirlines {
		log.Printf("      %d. %s: $%.0f", i+1, item.airline, item.price)
	}

	averagePrice := totalPrice / float64(validAirlines)

	// 計算價格範圍和標準差
	minPrice := sortedAirlines[0].price
	maxPrice := sortedAirlines[len(sortedAirlines)-1].price
	priceRange := maxPrice - minPrice

	// 計算標準差
	variance := 0.0
	for _, item := range sortedAirlines {
		variance += math.Pow(item.price-averagePrice, 2)
	}
	stdDev := math.Sqrt(variance / float64(validAirlines))

	log.Printf("   🎯 平均價格: $%.0f (基於 %d 個獨特航空公司)", averagePrice, validAirlines)
	log.Printf("   📊 價格範圍: $%.0f - $%.0f (範圍: $%.0f)", minPrice, maxPrice, priceRange)
	log.Printf("   📐 標準差: $%.0f", stdDev)

	return averagePrice, nil
}

// 輔助函數：檢查 slice 是否包含某個元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 新增：智能價格估算（當 API 失敗時使用）
func (s *AmadeusService) estimatePrice(origin, destination string, date time.Time, week, totalWeeks int) float64 {
	basePrice := s.getBasePrice(origin, destination)
	seasonalFactor := s.getSeasonalFactor(date)
	advanceDiscount := s.getAdvanceDiscount(week, totalWeeks)

	// 基於真實市場數據的估算
	estimatedPrice := basePrice * seasonalFactor * advanceDiscount

	log.Printf("   📊 智能估算價格: $%.0f (基礎: $%.0f, 季節: %.2f, 折扣: %.2f)",
		estimatedPrice, basePrice, seasonalFactor, advanceDiscount)

	return math.Round(estimatedPrice)
}

// 新增：獲取真實航班價格
func (s *AmadeusService) getRealFlightPrice(origin, destination string, date time.Time) (float64, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return 0, err
	}

	apiURL := fmt.Sprintf("%s/shopping/flight-offers", s.config.AmadeusBaseURL)

	params := url.Values{}
	params.Add("originLocationCode", origin)
	params.Add("destinationLocationCode", destination)
	params.Add("departureDate", date.Format("2006-01-02"))
	params.Add("adults", "1")
	params.Add("currencyCode", "TWD")
	params.Add("max", "5")

	fullURL := apiURL + "?" + params.Encode()

	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return 0, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API錯誤: 狀態碼 %d", resp.StatusCode)
	}

	// 解析響應獲取最低價格
	var apiResponse models.AmadeusFlightOffersResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return 0, err
	}

	if len(apiResponse.Data) == 0 {
		return 0, fmt.Errorf("未找到航班")
	}

	// 返回最低價格
	minPrice := 0.0
	for i, offer := range apiResponse.Data {
		price, _ := strconv.ParseFloat(offer.Price.Total, 64)
		if i == 0 || price < minPrice {
			minPrice = price
		}
	}

	return minPrice, nil
}

// 獲取航空公司名稱
func (s *AmadeusService) getAirlineName(code string) string {
	airlines := map[string]string{
		"CI": "中華航空",
		"BR": "長榮航空",
		"CX": "國泰航空",
		"JL": "日本航空",
		"NH": "全日空",
		"KE": "大韓航空",
		"SQ": "新加坡航空",
		"TG": "泰國航空",
		"UA": "聯合航空",
		"AA": "美國航空",
		"SL": "泰獅航空",
		"TW": "德威航空",
		"TR": "台灣虎航",
		"7C": "濟州航空",
		"MF": "廈門航空",
		"OZ": "韓亞航空",
		"JX": "星宇航空",
		"VN": "越南航空",
		"PR": "菲律賓航空",
	}

	if name, exists := airlines[code]; exists {
		return name
	}
	return code
}

// 新增：生成模擬價格（當真實API不可用時）
func (s *AmadeusService) generateMockPrice(origin, destination string, date time.Time, week, totalWeeks int) float64 {
	// 基礎價格（根據航線設定）
	basePrice := s.getBasePrice(origin, destination)

	// 季節性因素
	seasonalFactor := s.getSeasonalFactor(date)

	// 提前預訂折扣
	advanceDiscount := s.getAdvanceDiscount(week, totalWeeks)

	// 隨機波動
	rand.Seed(time.Now().UnixNano()) // 確保每次運行隨機數不同
	randomFactor := 0.85 + rand.Float64()*0.3

	// 週期性波動（模擬價格波動）
	weeklyPattern := 1.0 + 0.1*math.Sin(float64(week)*0.5)

	price := basePrice * seasonalFactor * advanceDiscount * randomFactor * weeklyPattern

	// 四捨五入到整數
	return math.Round(price)
}

// 新增：根據航線獲取基礎價格
func (s *AmadeusService) getBasePrice(origin, destination string) float64 {
	routePrices := map[string]float64{
		"TPE-TYO": 8000, // 台北-東京
		"TPE-OSA": 7500, // 台北-大阪
		"TPE-SEL": 6000, // 台北-首爾
		"TPE-HKG": 4000, // 台北-香港
		"TPE-BKK": 7000, // 台北-曼谷
		"TPE-SIN": 8000, // 台北-新加坡
		"TPE-KHH": 2000, // 台北-高雄（國內）
	}

	route := fmt.Sprintf("%s-%s", origin, destination)
	if price, exists := routePrices[route]; exists {
		return price
	}

	// 預設價格
	return 5000
}

// 新增：季節性因素
func (s *AmadeusService) getSeasonalFactor(date time.Time) float64 {
	month := date.Month()

	// 旺季（暑假、寒假、櫻花季等）
	switch month {
	case 1, 2: // 寒假、春節
		return 1.4
	case 3, 4: // 櫻花季
		return 1.3
	case 7, 8: // 暑假
		return 1.5
	case 12: // 聖誕節、跨年
		return 1.4
	default:
		return 1.0
	}
}

// 新增：提前預訂折扣
func (s *AmadeusService) getAdvanceDiscount(week, totalWeeks int) float64 {
	// 越早訂越便宜
	advanceRatio := float64(week) / float64(totalWeeks)

	if advanceRatio < 0.2 { // 前20%時間
		return 0.7 // 7折
	} else if advanceRatio < 0.5 { // 20%-50%時間
		return 0.8 // 8折
	} else if advanceRatio < 0.8 { // 50%-80%時間
		return 0.9 // 9折
	} else { // 最後20%時間
		return 1.0 // 原價
	}
}

// 新增：計算價格統計數據
// 修改：修復最佳日期計算邏輯
func (s *AmadeusService) calculatePriceStatistics(analysis *models.PriceAnalysis) {
	if len(analysis.DataPoints) == 0 {
		return
	}

	// 初始化為第一個數據點的值
	minPrice := analysis.DataPoints[0].Price
	maxPrice := analysis.DataPoints[0].Price
	sum := 0.0
	bestDate := analysis.DataPoints[0].Date // 初始化最佳日期

	log.Printf("📊 開始計算價格統計，共 %d 個數據點", len(analysis.DataPoints))

	for _, point := range analysis.DataPoints {
		log.Printf("   📅 第 %d 週: 日期=%s, 價格=$%.0f",
			point.Week, point.Date.Format("2006-01-02"), point.Price)

		if point.Price < minPrice {
			minPrice = point.Price
			bestDate = point.Date // 更新最佳日期
			log.Printf("   🎯 發現新的最低價格: $%.0f, 日期: %s", minPrice, bestDate.Format("2006-01-02"))
		}
		if point.Price > maxPrice {
			maxPrice = point.Price
		}
		sum += point.Price
	}

	analysis.MinPrice = minPrice
	analysis.MaxPrice = maxPrice
	analysis.AvgPrice = sum / float64(len(analysis.DataPoints))
	analysis.BestDate = bestDate // 設置最佳日期
	analysis.Recommendation = s.generateRecommendation(analysis)

	log.Printf("✅ 統計計算完成:")
	log.Printf("   📈 最低價格: $%.0f", analysis.MinPrice)
	log.Printf("   📈 最高價格: $%.0f", analysis.MaxPrice)
	log.Printf("   📊 平均價格: $%.0f", analysis.AvgPrice)
	log.Printf("   🎯 最佳出發日期: %s", analysis.BestDate.Format("2006-01-02"))
	log.Printf("   💡 推薦建議: %s", analysis.Recommendation)
}

// 新增：生成推薦建議
func (s *AmadeusService) generateRecommendation(analysis *models.PriceAnalysis) string {
	savings := analysis.AvgPrice - analysis.MinPrice
	savingsRatio := (savings / analysis.AvgPrice) * 100

	bestDateStr := analysis.BestDate.Format("2006年1月2日")

	if savingsRatio > 20 {
		return fmt.Sprintf("強烈建議在 %s 出發！價格 $%.0f 為最低價，相比平均價格節省 $%.0f (%.0f%%)",
			bestDateStr, analysis.MinPrice, savings, savingsRatio)
	} else if savingsRatio > 10 {
		return fmt.Sprintf("建議在 %s 出發，價格 $%.0f 較為優惠，可節省 $%.0f",
			bestDateStr, analysis.MinPrice, savings)
	} else {
		return "價格波動不大，可根據個人行程安排選擇出發時間"
	}
}

// 新增：生成價格趨勢數據（用於圖表）
func (s *AmadeusService) GeneratePriceTrend(origin, destination string, weeks int) (*models.PriceTrend, error) {
	route := fmt.Sprintf("%s-%s", origin, destination)

	s.trackingMutex.RLock()
	analysis, exists := s.trackingData[route]
	s.trackingMutex.RUnlock()

	if !exists {
		// 如果沒有歷史數據，創建新的追蹤
		req := models.PriceTrackingRequest{
			Origin:      origin,
			Destination: destination,
			Weeks:       weeks,
		}
		var err error
		analysis, err = s.TrackFlightPrices(req)
		if err != nil {
			return nil, err
		}
	}

	// 轉換為圖表數據格式
	trend := &models.PriceTrend{
		Route:   route,
		Weeks:   weeks,
		Summary: analysis,
	}

	// 準備圖表數據
	for _, point := range analysis.DataPoints {
		trend.Labels = append(trend.Labels, point.Date.Format("01/02"))
		trend.Prices = append(trend.Prices, point.Price)
		trend.WeekNums = append(trend.WeekNums, point.Week) // 改為 WeekNums
	}

	return trend, nil
}

// 原有的方法保持不變...
// 獲取訪問令牌
func (s *AmadeusService) getAccessToken() (string, error) {
	// 如果令牌還有至少5分鐘有效期，直接返回
	if s.accessToken != "" && time.Now().Before(s.tokenExpiry.Add(-5*time.Minute)) {
		return s.accessToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.config.AmadeusAPIKey)
	data.Set("client_secret", s.config.AmadeusAPISecret)

	req, err := http.NewRequest("POST", "https://test.api.amadeus.com/v1/security/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("創建令牌請求失敗: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("令牌請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("讀取令牌響應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("令牌獲取失敗: %s", string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("解析令牌響應失敗: %v", err)
	}

	s.accessToken = tokenResponse.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)

	log.Println("✅ Amadeus訪問令牌獲取成功")
	return s.accessToken, nil
}

// 搜尋航班報價
func (s *AmadeusService) SearchFlights(req models.SearchRequest) ([]models.Flight, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	// 構建API URL
	apiURL := fmt.Sprintf("%s/shopping/flight-offers", s.config.AmadeusBaseURL)

	// 構建查詢參數
	params := url.Values{}
	params.Add("originLocationCode", req.Origin)
	params.Add("destinationLocationCode", req.Destination)
	params.Add("departureDate", req.DepartureDate)

	if req.ReturnDate != "" {
		params.Add("returnDate", req.ReturnDate)
	}

	params.Add("adults", strconv.Itoa(req.Adults))
	params.Add("currencyCode", req.Currency)
	params.Add("max", "10") // 限制結果數量

	fullURL := apiURL + "?" + params.Encode()

	// 創建請求
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("創建請求失敗: %v", err)
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	log.Printf("🔍 搜尋航班: %s -> %s 日期: %s", req.Origin, req.Destination, req.DepartureDate)

	// 發送請求
	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API請求失敗: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("讀取響應失敗: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ API錯誤響應: %s", string(body))
		return nil, fmt.Errorf("API錯誤: 狀態碼 %d", resp.StatusCode)
	}

	// 解析響應
	var apiResponse models.AmadeusFlightOffersResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("解析JSON失敗: %v", err)
	}

	log.Printf("✅ 找到 %d 個航班報價", len(apiResponse.Data))

	// 轉換為統一格式
	return s.transformResponse(apiResponse), nil
}

// 轉換Amadeus響應為統一格式
func (s *AmadeusService) transformResponse(response models.AmadeusFlightOffersResponse) []models.Flight {
	var flights []models.Flight

	for _, offer := range response.Data {
		if len(offer.Itineraries) == 0 {
			continue
		}

		itinerary := offer.Itineraries[0]
		if len(itinerary.Segments) == 0 {
			continue
		}

		firstSegment := itinerary.Segments[0]
		lastSegment := itinerary.Segments[len(itinerary.Segments)-1]

		// 解析價格
		price, _ := strconv.ParseFloat(offer.Price.Total, 64)

		// [修改] 不再解析時間，直接使用 API 回傳的原始字串
		// 這樣可以保留 "當地時間" 語意，避免時區轉換錯誤導致的 08:06 問題

		// 獲取航空公司名稱
		airline := s.getAirlineName(firstSegment.CarrierCode)

		flight := models.Flight{
			ID:           offer.ID,
			Price:        price,
			Currency:     offer.Price.Currency,
			Airline:      airline,
			FlightNumber: fmt.Sprintf("%s%s", firstSegment.CarrierCode, firstSegment.Number),
			From: models.Airport{
				Code:     firstSegment.Departure.IATACode,
				Terminal: firstSegment.Departure.Terminal,
			},
			To: models.Airport{
				Code:     lastSegment.Arrival.IATACode,
				Terminal: lastSegment.Arrival.Terminal,
			},
			// [修改] 直接賦值字串
			Departure: firstSegment.Departure.At,
			Arrival:   lastSegment.Arrival.At,
			Duration:  itinerary.Duration,
			Stops:     len(itinerary.Segments) - 1,
			Aircraft:  firstSegment.Aircraft.Code,
		}

		flights = append(flights, flight)
	}

	return flights
}

// 搜尋機場
func (s *AmadeusService) SearchAirports(keyword string) ([]models.Airport, error) {
	token, err := s.getAccessToken()
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/reference-data/locations", s.config.AmadeusBaseURL)

	params := url.Values{}
	params.Add("subType", "AIRPORT")
	params.Add("keyword", keyword)
	params.Add("page[limit]", "10")

	fullURL := apiURL + "?" + params.Encode()

	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Add("Authorization", "Bearer "+token)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("機場搜尋失敗: 狀態碼 %d", resp.StatusCode)
	}

	var apiResponse models.AirportResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	var airports []models.Airport
	for _, airport := range apiResponse.Data {
		airports = append(airports, models.Airport{
			Code: airport.IATACode,
			Name: airport.Name,
			City: airport.Address.CityName,
		})
	}

	return airports, nil
}
