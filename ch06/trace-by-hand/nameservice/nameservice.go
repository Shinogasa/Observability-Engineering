package nameservice

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// NameHandler は名前サービスのHTTPハンドラー
// 受信したB3ヘッダーを使って自分自身のスパンを生成する
func NameHandler(w http.ResponseWriter, r *http.Request) {
	// 親スパン情報をB3ヘッダーから取得
	traceID := r.Header.Get("X-B3-TraceId")
	parentSpanID := r.Header.Get("X-B3-ParentSpanId")

	// このサービス自身のスパンを生成
	traceData := make(map[string]interface{})
	traceData["trace_id"] = traceID
	traceData["span_id"] = uuid.New().String()
	traceData["parent_span_id"] = parentSpanID
	traceData["name"] = "/name"
	traceData["service_name"] = "nameservice"

	startTime := time.Now()
	traceData["timestamp"] = startTime.Unix()

	// 名前取得ロジック（モック: 固定値を返す）
	name := "Charity"

	// JSONレスポンス
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"name": name})

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
