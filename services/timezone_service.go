package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TimeZoneResponse struct {
	TimeZone  string `json:"timezone"`
	UTCOffset string `json:"utc_offset"`
	Datetime  string `json:"datetime"`
}

func GetTimeZone(location string) (TimeZoneResponse, error) {
	// WorldTimeAPI 格式: http://worldtimeapi.org/api/timezone/Asia/Taipei
	url := fmt.Sprintf("https://worldtimeapi.org/api/timezone/%s", location)

	resp, err := http.Get(url)
	if err != nil {
		return TimeZoneResponse{}, err
	}
	defer resp.Body.Close()

	var result TimeZoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return TimeZoneResponse{}, err
	}
	return result, nil
}

// 計算兩個地點的時差
func CalculateTimeDifference(loc1, loc2 string) (float64, error) {
	tz1, err := GetTimeZone(loc1)
	if err != nil {
		return 0, err
	}
	tz2, err := GetTimeZone(loc2)
	if err != nil {
		return 0, err
	}

	// 解析時間字串
	t1, err := time.Parse(time.RFC3339, tz1.Datetime)
	if err != nil {
		return 0, err
	}
	t2, err := time.Parse(time.RFC3339, tz2.Datetime)
	if err != nil {
		return 0, err
	}

	diff := t2.Sub(t1).Hours()
	return diff, nil
}
