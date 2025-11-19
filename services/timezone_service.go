package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TimeZoneResponse struct {
	TimeZone  string `json:"timezone"`
	UTCOffset string `json:"utc_offset"`
	Datetime  string `json:"datetime"`
}

// 緩存時區計算結果
var (
	timezoneCache = make(map[string]TimeZoneResponse)
	cacheMutex    = &sync.RWMutex{}
)

// 使用系統內建時區資料（快速、離線）
func GetTimeZone(location string) (TimeZoneResponse, error) {
	// 先檢查緩存
	cacheMutex.RLock()
	if cached, exists := timezoneCache[location]; exists {
		cacheMutex.RUnlock()
		fmt.Printf("⚡ 從緩存取得時區: %s\n", location)
		return cached, nil
	}
	cacheMutex.RUnlock()

	fmt.Printf("🔍 載入系統時區: %s\n", location)

	// 載入時區
	loc, err := time.LoadLocation(location)
	if err != nil {
		return TimeZoneResponse{}, fmt.Errorf("時區 '%s' 不存在，請使用 Region/City 格式，例如: Asia/Taipei, Europe/London, America/New_York", location)
	}

	// 取得當前時間在該時區
	now := time.Now().In(loc)

	// 計算 UTC 偏移量
	_, offset := now.Zone()
	offsetHours := float64(offset) / 3600.0

	// 格式化 UTC 偏移量
	utcOffset := formatUTCOffset(offsetHours)

	response := TimeZoneResponse{
		TimeZone:  location,
		UTCOffset: utcOffset,
		Datetime:  now.Format(time.RFC3339),
	}

	// 存入緩存
	cacheMutex.Lock()
	timezoneCache[location] = response
	cacheMutex.Unlock()

	fmt.Printf("✅ 系統時區資訊: %s (UTC%s)\n", location, utcOffset)

	return response, nil
}

// 計算兩個地點的時差
func CalculateTimeDifference(loc1, loc2 string) (float64, error) {
	fmt.Printf("⏰ 計算時差: %s → %s\n", loc1, loc2)

	tz1, err := GetTimeZone(loc1)
	if err != nil {
		return 0, fmt.Errorf("無法取得時區 '%s': %v", loc1, err)
	}
	tz2, err := GetTimeZone(loc2)
	if err != nil {
		return 0, fmt.Errorf("無法取得時區 '%s': %v", loc2, err)
	}

	// 直接比較 UTC 偏移量
	offset1, err := parseUTCOffset(tz1.UTCOffset)
	if err != nil {
		return 0, fmt.Errorf("無法解析時區 '%s' 的 UTC 偏移量: %v", loc1, err)
	}

	offset2, err := parseUTCOffset(tz2.UTCOffset)
	if err != nil {
		return 0, fmt.Errorf("無法解析時區 '%s' 的 UTC 偏移量: %v", loc2, err)
	}

	// 時差 = 目標時區偏移量 - 起始時區偏移量
	diff := offset2 - offset1

	fmt.Printf("🎯 時差計算結果: %s (UTC%s) → %s (UTC%s) = %.1f 小時\n",
		loc1, tz1.UTCOffset, loc2, tz2.UTCOffset, diff)

	return diff, nil
}

// 格式化 UTC 偏移量
func formatUTCOffset(offsetHours float64) string {
	hours := int(offsetHours)
	minutes := int((offsetHours - float64(hours)) * 60)
	if minutes < 0 {
		minutes = -minutes
	}

	sign := "+"
	if hours < 0 {
		sign = "-"
		hours = -hours
	}

	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

// 解析 UTC 偏移量字串 (例如: "+08:00", "-05:00")
func parseUTCOffset(offsetStr string) (float64, error) {
	if offsetStr == "" {
		return 0, fmt.Errorf("UTC 偏移量為空")
	}

	// 移除可能的空格
	offsetStr = strings.TrimSpace(offsetStr)

	// 檢查格式
	if len(offsetStr) < 6 || (offsetStr[0] != '+' && offsetStr[0] != '-') {
		return 0, fmt.Errorf("無效的 UTC 偏移量格式: %s", offsetStr)
	}

	// 分割小時和分鐘
	parts := strings.Split(offsetStr[1:], ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("無效的 UTC 偏移量格式: %s", offsetStr)
	}

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("無法解析小時: %v", err)
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, fmt.Errorf("無法解析分鐘: %v", err)
	}

	// 計算總小時數（包含正負號）
	totalHours := float64(hours) + float64(minutes)/60.0
	if offsetStr[0] == '-' {
		totalHours = -totalHours
	}

	return totalHours, nil
}

// 取得支援的時區列表（用於前端自動完成）
func GetSupportedTimeZones() []string {
	return []string{
		"Asia/Taipei",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Seoul",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"America/New_York",
		"America/Los_Angeles",
		"America/Chicago",
		"Australia/Sydney",
		"Australia/Melbourne",
	}
}
