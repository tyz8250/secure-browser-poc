# secure-browser-poc

自治体の三層分離環境におけるSaaS利用を、**使い捨ての隔離ブラウザ環境で改善できないか検証するPoC**です。

## Why

自治体の三層分離環境では、LGWAN接続端末から一般のクラウドサービスを利用する場合、インターネット接続環境へ切り替える必要があります。

そこで、端末自体をインターネットへ接続するのではなく、必要なときだけ隔離された一時環境を起動し、そこからSaaSへアクセスできないかと考えました。

利用終了後に環境を破棄することで、**隔離・通信制御・破棄を前提としたSaaS利用の構成が成立するか**を検証しています。

Dockerを使うこと自体が目的ではなく、一時環境をAPIから生成・管理・破棄する仕組みを小さく検証するため、現在はDockerコンテナを利用しています。

## Current Status

🚧 開発中（PoC）

現在は、隔離環境の基本的なライフサイクル管理を実装しています。

## Implemented

- `POST /sessions`
  - セッションを作成
  - Dockerコンテナを起動
- `GET /sessions/{id}`
  - コンテナの状態を取得
- `DELETE /sessions/{id}`
  - コンテナを停止・削除
- 基本的なエラー処理

## Tech Stack

- Go
- Docker
- Docker Go SDK
- HTTP API

## Next

- timeout / cleanup
- network isolation
- egress control / proxy
- logs / metrics
- resource limits
- failure handling

## Concept

```text
User
  ↓
Gateway
  ↓
Session Manager
  ↓
Temporary Browser Environment
  ↓
Allowed SaaS

利用終了
  ↓
Session Manager
  ↓
環境を破棄
```

最終的には、利用者がDockerなどの基盤技術を意識せず、必要なときだけ隔離環境を利用できる仕組みを目指しています。