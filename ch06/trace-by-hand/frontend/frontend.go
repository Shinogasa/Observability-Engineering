package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

const (
	authServiceURL = "http://localhost:8081/auth"
	nameServiceURL = "http://localhost:8082/name"
)

// RootHandler は書籍6.3節の rootHandler に対応する
// トレースIDとスパンIDを生成し、下流サービスへB3ヘッダーで伝搬する
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// トレースデータの初期化
	traceData := make(map[string]interface{})
	traceData["tags"] = make(map[string]interface{})

	hostname, _ := os.Hostname()
	traceData["tags"].(map[string]interface{})["hostname"] = hostname

	traceData["trace_id"] = uuid.New().String()
	traceData["span_id"] = uuid.New().String()

	// タイムスタンプの記録
	startTime := time.Now()
	traceData["timestamp"] = startTime.Unix()

	// スパンの名前とサービス名を設定
	traceData["name"] = "/root"
	traceData["service_name"] = "frontend"

	// 下流サービスの呼び出し（B3ヘッダーでトレース情報を伝搬）
	authorized := callAuthService(r, traceData)
	name := callNameService(r, traceData)

	// カスタムフィールドの追加（書籍6.4節）
	traceData["tags"].(map[string]interface{})["user_name"] = name

	// レスポンス生成
	w.Header().Set("Content-Type", "application/json")
	if authorized {
		w.Write([]byte(fmt.Sprintf(`{"message": "Waddup %s"}`, name)))
	} else {
		w.Write([]byte(`{"message": "Not cool dawg"}`))
	}

	// スパンの完了と送信
	traceData["duration_ms"] = time.Since(startTime).Milliseconds()
	sendSpan(traceData)
}

// callAuthService は認証サービスを呼び出し、B3ヘッダーでトレース情報を伝搬する
func callAuthService(r *http.Request, traceData map[string]interface{}) bool {
	req, err := http.NewRequest("GET", authServiceURL, nil)
	if err != nil {
		return false
	}

	// B3トレースヘッダーの設定
	req.Header.Set("X-B3-TraceId", traceData["trace_id"].(string))
	req.Header.Set("X-B3-ParentSpanId", traceData["span_id"].(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// callNameService は名前サービスを呼び出し、B3ヘッダーでトレース情報を伝搬する
func callNameService(r *http.Request, traceData map[string]interface{}) string {
	req, err := http.NewRequest("GET", nameServiceURL, nil)
	if err != nil {
		return "unknown"
	}

	// B3トレースヘッダーの設定
	req.Header.Set("X-B3-TraceId", traceData["trace_id"].(string))
	req.Header.Set("X-B3-ParentSpanId", traceData["span_id"].(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "unknown"
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "unknown"
	}
	return result["name"]
}

// sendSpan はトレースデータをJSON形式で標準出力に出力する
// 本番ではJaeger/Zipkin等のトレースバックエンドへ送信する
func sendSpan(traceData map[string]interface{}) {
	data, err := json.Marshal(traceData)
	if err != nil {
		return
	}
	fmt.Println(string(data))
}
