package authservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// AuthHandler は認証サービスのHTTPハンドラー
// 受信したB3ヘッダーを使って自分自身のスパンを生成する
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	// 親スパン情報をB3ヘッダーから取得
	traceID := r.Header.Get("X-B3-TraceId")
	parentSpanID := r.Header.Get("X-B3-ParentSpanId")

	// このサービス自身のスパンを生成
	traceData := make(map[string]interface{})
	traceData["trace_id"] = traceID            // 伝搬されたtrace_idをそのまま使用
	traceData["span_id"] = uuid.New().String()  // 新しいspan_idを生成
	traceData["parent_span_id"] = parentSpanID  // frontendのspan_idを親として記録
	traceData["name"] = "/auth"
	traceData["service_name"] = "authservice"

	startTime := time.Now()
	traceData["timestamp"] = startTime.Unix()

	// 認証ロジック（モック: 常に認証OK）
	authorized := true

	if authorized {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}

	// スパンの完了と送信
	traceData["duration_ms"] = time.Since(startTime).Milliseconds()
	sendSpan(traceData)
}

// sendSpan はトレースデータをJSON形式で標準出力に出力する
func sendSpan(traceData map[string]interface{}) {
	data, err := json.Marshal(traceData)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}
