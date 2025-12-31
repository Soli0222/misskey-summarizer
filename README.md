# Misskey Daily Summary CLI

Misskeyで投稿した1日分のノートを取得し、OpenAI APIを使用してその日のサマリーを生成するCLIアプリケーションです。

## 機能

- 📅 指定した日付のノートを取得
- 🤖 OpenAI (GPT-5 Mini) を使用してその日の活動をサマリー
- 💬 Discordにサマリーを投稿
- 🔄 Renote（リノート）も含めて分析
- ☸️ Kubernetes CronJobで定期実行可能

## プロジェクト構造

[Standard Go Project Layout](https://github.com/golang-standards/project-layout) に準拠しています。

```
.
├── cmd/
│   └── misskey-summarizer/    # アプリケーションエントリーポイント
│       ├── main.go
│       └── cmd/
│           └── root.go        # Cobraルートコマンド
├── internal/                  # プライベートパッケージ
│   ├── misskey/               # Misskey APIクライアント
│   │   ├── client.go
│   │   ├── aid.go             # AIDアルゴリズム
│   │   └── aid_test.go
│   ├── openai/                # OpenAI APIクライアント
│   │   └── client.go
│   ├── discord/               # Discord Webhook
│   │   └── client.go
│   └── models/                # データモデル
│       └── note.go
├── k8s/                       # Kubernetes マニフェスト
│   ├── cronjob.yaml
│   └── secret.example.yaml
├── Dockerfile
├── Makefile
└── README.md
```

## インストール

```bash
go install github.com/soli0222/misskey-summarizer/cmd/misskey-summarizer@latest
```

または、リポジトリをクローンしてビルド:

```bash
git clone https://github.com/soli0222/misskey-summarizer.git
cd misskey-summarizer
make build
```

## 設定

以下の環境変数を設定してください:

| 環境変数 | 説明 | 必須 |
|----------|------|------|
| `MISSKEY_TOKEN` | Misskey APIトークン | ✅ |
| `MISSKEY_INSTANCE_URL` | MisskeyインスタンスのURL | ✅ |
| `OPENAI_API_KEY` | OpenAI APIキー | ✅ |
| `OPENAI_MODEL` | 使用するAIモデル | ❌ (デフォルト: `gpt-5-mini-2025-08-07`) |
| `DISCORD_WEBHOOK_URL` | Discord Webhook URL | ❌ (`--discord`使用時に必要) |

## 使い方

### 基本的な使い方

```bash
# 今日のノートをサマリー
./misskey-summarizer

# 昨日のノートをサマリー（CronJob向け）
./misskey-summarizer --yesterday

# 特定の日付を指定
./misskey-summarizer --date 2025-12-30

# JSON形式で出力
./misskey-summarizer --yesterday --output json

# Discordに投稿
./misskey-summarizer --yesterday --discord
```

### コマンドラインオプション

```
Usage:
  misskey-summarizer [flags]

Flags:
  -d, --date string     Target date in YYYY-MM-DD format
      --discord         Post summary to Discord webhook
  -h, --help            help for misskey-summarizer
  -o, --output string   Output format: summary or json (default "summary")
  -y, --yesterday       Use yesterday's date
```

## Kubernetes CronJob での運用

### Secretの作成

```bash
kubectl create secret generic misskey-summary-secrets \
  --from-literal=misskey-token=YOUR_TOKEN \
  --from-literal=openai-api-key=YOUR_API_KEY \
  --from-literal=discord-webhook-url=YOUR_WEBHOOK_URL
```

### CronJobのデプロイ

```bash
kubectl apply -f k8s/cronjob.yaml
```

### 手動実行

```bash
kubectl create job --from=cronjob/misskey-summary misskey-summary-manual
```

## Docker

### イメージのビルド

```bash
make docker-build
```

### Dockerで実行

```bash
docker run --rm \
  -e MISSKEY_TOKEN=your-token \
  -e MISSKEY_INSTANCE_URL=https://mi.soli0222.com \
  -e OPENAI_API_KEY=your-api-key \
  -e DISCORD_WEBHOOK_URL=your-webhook-url \
  ghcr.io/soli0222/misskey-summarizer:latest --yesterday --discord
```

## 開発

```bash
# 依存関係のインストール
go mod tidy

# ビルド
make build

# テスト
make test

# フォーマット
make fmt
```

## ライセンス

MIT
