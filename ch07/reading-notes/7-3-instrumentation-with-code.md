# 7.3 コードベースの例を用いた計装

## 7.3.2 カスタム計装を追加する

自動計装（HTTPやgRPCのインターセプター等）で基盤が整ったら、ビジネスロジックに特化したカスタム計装に投資できる。

### トレーススパンの開始と終了

- トレーススパンは実行された個々の作業単位をカバーする
- 単位は通常1つのマイクロサービスを通過する個々のリクエストで、HTTPリクエストまたはRPCレイヤーでインスタンス化する
- より細かい粒度が必要な場合、カスタムスパンを追加してプロセス内部の可視化が可能

```go
import "go.opentelemetry.io/otel"

// 各モジュール内で識別しやすくするよう、一度だけプライベートトレーサーを定義
var tr = otel.Tracer("module_name")

func funcName(ctx context.Context) {
    sp := tr.Start(ctx, "span_name")
    defer sp.End()
    // 何らかの処理
}
```

- `tr.Start(ctx, "span_name")` でコンテキスト内の既存スパンから子スパンを派生
- `defer sp.End()` で関数終了時にスパンの持続時間を計算し、バックエンドに送信

### イベントに幅広いフィールドを追加する

- 自動計装されたスパンに、クライアントID・シャードID・エラーなどのカスタムフィールドやリッチな値を添付できる
- 現在アクティブなスパンは `sp := trace.SpanFromContext(ctx)` で取得可能

```go
import "go.opentelemetry.io/otel/attribute"

sp.SetAttributes(attribute.Int("http.code", resp.ResponseCode))
sp.SetAttributes(attribute.String("app.user", username))
```

- Goでは型安全性のため、`attribute.Int`、`attribute.String` など明示的に型を指定する

### プロセス全体のメトリクスを記録する

メトリクスの記録場所には2つのパターンがある。

#### パターン1: リクエストスコープ → スパンの属性として記録

特定のHTTPリクエストやRPC呼び出しに紐づく値（例: カートの合計金額、レスポンスコード）は、処理中のスパンに属性として付与するのが適切。

#### パターン2: プロセススコープ → メトリクスAPIで直接記録

特定のリクエストに属さない、プロセス全体の状態を示す値はメトリクスAPIで記録する。

- 例: 実行中のgoroutine数、メモリ使用量、コネクションプール数
- これらは「どのリクエストのスパンに付けるか？」に答えられない → スパンに付けるのは不適切
- 定期的に実行されるgoroutine（タイマー）内でメトリクスAPIを使って記録する

```go
import "go.opentelemetry.io/otel"
import "go.opentelemetry.io/otel/metric"

// トレーサーと同様に、パッケージごとにメーターが必要
var meter = otel.Meter("example_package")

// キーを事前に定義して再利用することで、オーバーヘッドを回避
var appKey = attribute.Key("app")
var containerKey = attribute.Key("container")

goroutines, _ := metric.Must(meter).NewInt64Measure("num_goroutines",
    metric.WithKeys(appKey, containerKey),
    metric.WithDescription("Amount of goroutines running."),
)

// 周期的に実行されるゴルーチン内で記録
meter.RecordBatch(ctx, []attribute.KeyValue{
    appKey.String(os.Getenv("PROJECT_DOMAIN")),
    containerKey.String(os.Getenv("HOSTNAME"))},
    goroutines.Measurement(int64(runtime.NumGoroutine())),
)
```

#### なぜキーを事前定義するとオーバーヘッドを回避できるのか

メトリクスの記録はホットパス（頻繁に実行されるコードパス）になりやすい。

**毎回インラインでキーを生成する場合:**

```go
// 記録のたびに attribute.Key オブジェクトを新規生成
attribute.String("app", os.Getenv("PROJECT_DOMAIN"))
```

- 呼び出しのたびにヒープにメモリアロケーションが発生する可能性がある
- 生成された一時オブジェクトはすぐ不要になり、GC（ガベージコレクション）が回収する必要がある
- 高頻度な記録では、小さなアロケーションの累積がGCの停止時間を増大させる

**事前に定義して再利用する場合:**

```go
var appKey = attribute.Key("app")  // 起動時に1回だけアロケーション
appKey.String(os.Getenv("PROJECT_DOMAIN"))  // 既存のキーを再利用
```

- プロセス起動時に1回だけアロケーションが発生
- 以降は同じオブジェクトを使い回すためアロケーションゼロ
- GCへの負荷も最小限

**補足:** 計装コードはビジネスロジックのあらゆる場所に埋め込まれるため、計装自体がパフォーマンスボトルネックにならないよう気を配ることが重要。「ホットパスでのアロケーションを減らす」はGo（や他の言語）のパフォーマンス最適化の基本原則。
