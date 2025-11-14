// handlers/timezone.go (修正版 - 專門處理 API 請求)
package handlers

import (
	"encoding/json"
	"final/services"
	"net/http"
	"strconv" // 新增
)

// TimeDiffHandler 處理時差計算的 POST API 請求
func TimeDiffHandler(w http.ResponseWriter, r *http.Request) {
	// 只處理 POST 請求
	if r.Method != "POST" {
		http.Error(w, "僅允許 POST 請求", http.StatusMethodNotAllowed)
		return
	}

	// 解析表單數據
	if err := r.ParseForm(); err != nil {
		http.Error(w, "無法解析表單數據", http.StatusBadRequest)
		return
	}

	from := r.FormValue("from")
	to := r.FormValue("to")

	// 確保時區輸入不為空
	if from == "" || to == "" {
		http.Error(w, "請提供起始和目標時區", http.StatusBadRequest)
		return
	}

	// 呼叫 Service 層計算時差
	diff, err := services.CalculateTimeDifference(from, to)
	if err != nil {
		// 伺服器錯誤 (例如 API 連線失敗或時區名稱錯誤)
		http.Error(w, "計算時差失敗: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 準備 JSON 響應
	w.Header().Set("Content-Type", "application/json")
	// 格式化輸出的小時數（例如：8.5 或 -7.0）
	response := map[string]interface{}{
		"success": true,
		"from":    from,
		"to":      to,
		// 使用 strconv.FormatFloat 來格式化為一位小數的字串
		"diff":    diff,
		"diffStr": strconv.FormatFloat(diff, 'f', 1, 64) + " 小時",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "無法編碼 JSON 響應", http.StatusInternalServerError)
		return
	}
}
