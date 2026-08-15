# LinkVault

[![build](https://github.com/matsu0122-png/linkvault/actions/workflows/build.yaml/badge.svg)](https://github.com/matsu0122-png/linkvault/actions/workflows/build.yaml)

[![Coverage Status](https://coveralls.io/repos/github/matsu0122-png/linkvault/badge.svg?branch=main)](https://coveralls.io/github/matsu0122-png/linkvault?branch=main)

> Webで見つけた大切な情報を、あとから簡単に見つけられるように。



![Status](https://img.shields.io/badge/status-development-orange)

## 概要

LinkVaultは、Web上で見つけた記事、技術ドキュメント、GitHubリポジトリなどのURLを保存・整理し、あとから簡単に探せるようにするWebアプリケーションです。

Web上で有用な情報を見つけても、時間が経つと「どこで見つけたのか」「何のために保存したのか」が分からなくなることがあります。

LinkVaultは、単にURLを貯めるだけのブックマークではなく、**「保存は一瞬、整理は自動、必要になったらすぐ見つかる」**を目指しています。URLを貼るだけでページのタイトルを自動取得し、タグとメモを添えて、あとから検索・絞り込みで一瞬に見つけ出せるようにします。

## 主な機能

* URLを入力するだけでリンクを保存（タイトル・説明文・アイコンをWebページから自動取得）
* Chrome拡張から、今見ているページをワンクリックで保存
* 複数URLのまとめて登録
* 登録したリンクの一覧表示
* リンクの編集・削除
* タグによる整理・絞り込み
* メモの保存
* キーワードによる検索
* 保存したURLの生存確認・リンク切れの検出
* 大量のリンクをページ単位で閲覧・並び替え
* Collection（コレクション）による関連リンクのグループ管理

## 使い方

保存したいWebページのURLを入力するだけで登録できます。タイトルを空欄のままにすると、そのページの`<title>`や説明文（OGP）、アイコン（favicon）を自動で取得します。

例：

```text
URL:
https://example.com/go-concurrency

タイトル:
（空欄のままでOK。Go Concurrency Patterns のように自動で補完されます）

タグ:
Go, goroutine, Backend

メモ:
Worker Poolについて学ぶ際の参考資料
```

登録したリンクはLinkVault上で一覧表示され、検索やタグによる絞り込みができます。

```text
LinkVault

検索: [ Go concurrency ]

---------------------------------

Go Concurrency Patterns

https://example.com/go-concurrency

#Go #goroutine #Backend

Worker Poolについて学ぶ際の参考資料
```

## Collection

タグとは別に、Collection（コレクション）で関連するリンクをひとまとまりにして管理できます。

* **Collection** → 「Go学習」「仕事」「お気に入り」のように、リンクをどのグループにまとめるか
* **Tag** → 「Go」「AWS」「Documentation」のように、リンクがどんな属性を持つか

1つのリンクは複数のCollectionへ同時に所属できます。サイドバーのCOLLECTIONSからCollectionを作成し、リンクの作成・編集フォームからチェックボックスで所属させたいCollectionを選べます。Collectionを削除してもリンク自体は削除されません。

CollectionはCollectionの中にさらにCollectionを作れます（入れ子構造）。サイドバーで各Collectionにマウスを乗せると表示される「＋」から、そのCollectionの子として新しいCollectionを作成できます。親Collectionを削除しても、中にあった子Collectionは削除されず最上位に繰り上がります（リンクと同じく、削除しても中身は消えない設計です）。

## テスト

### バックエンド

```bash
cd backend
go vet ./...
go test ./...
```

`repository`パッケージのテストは実際のPostgreSQL（`linkvault_test`という専用データベース）へ接続する統合テスト。接続できない場合は自動的にスキップされる。`service`・`handler`パッケージのテストはモックを使った単体テストで、DB接続は不要。

### フロントエンド

```bash
cd frontend
npm run lint
npm run build
```

現時点でフロントエンドに自動テスト（vitest等）は導入していない。動作確認はブラウザでの手動確認、および`npm run build`（型チェック含む）で行っている。

## Chrome拡張

`extension/`ディレクトリにMV3のChrome拡張機能があります。ビルド不要のプレーンなHTML/CSS/JSで、開いているページのURL・タイトルを取得してツールバーのポップアップからワンクリックで保存できます。

1. `chrome://extensions`を開き、「デベロッパーモード」を有効にする
2. 「パッケージ化されていない拡張機能を読み込む」から`extension/`ディレクトリを選択
3. 読み込み後に表示される拡張ID（`chrome-extension://...`）を、バックエンドの`CORS_ALLOWED_ORIGIN`環境変数にカンマ区切りで追加する（例: `CORS_ALLOWED_ORIGIN=http://localhost:5173,chrome-extension://<拡張ID>`）

## アーキテクチャ

```text
frontend/    React + TypeScript + Vite製のWebアプリ（http://localhost:5173）
backend/     Go製のREST API（http://localhost:8080）
             config / database / handler / service / repository / model の層構成
extension/   Chrome拡張（Manifest V3、ビルド不要のプレーンHTML/CSS/JS）
```

フロントエンド・Chrome拡張はいずれもbackendのREST APIをJSON経由で叩く。データベースはPostgreSQLで、`backend/migrations/`配下のSQLファイルを順に適用して構築する。詳細な設計・API仕様は[`.github/assets/spec.md`](.github/assets/spec.md)を参照。

## 使用技術

### フロントエンド

* React
* TypeScript
* Vite
* Tailwind CSS

### バックエンド

* Go
* 標準ライブラリの`net/http`（Webフレームワーク不使用）

### データベース

* PostgreSQL

### CI / テスト

* GitHub Actions
* Coveralls

## Docker

Docker Composeで、Frontend・Backend・PostgreSQLをまとめて起動できる。とりあえず動かして触ってみたい場合はこちらが最短。

### 必要なもの

* Docker
* Docker Compose

### セットアップ

```bash
cp .env.example .env
```

`.env`はGit管理しない（`.gitignore`に含めている）。`.env.example`の値はローカル開発用のダミー値で、本番の秘密情報ではない。

### 起動

```bash
docker compose up --build
```

* Frontend: http://localhost:5173
* Backend: http://localhost:8080

初回起動時、`backend/migrations/`配下のSQLファイルがPostgreSQLコンテナの初期化時に自動適用される（`docker-entrypoint-initdb.d`の仕組みを利用。既存のmigration方式をそのまま流用しており、Docker用の別方式は導入していない）。**この自動適用は「データディレクトリが空の初回起動時」のみ**。起動後に新しいmigrationファイルを追加した場合は自動適用されないため、後述の「完全削除」でVolumeごと作り直すか、`docker compose exec db psql -U $POSTGRES_USER -d $POSTGRES_DB -f /docker-entrypoint-initdb.d/000N_xxx.sql`で手動適用する。

### 基本操作

```bash
docker compose ps          # 各サービスの状態確認
docker compose logs -f     # ログを追跡表示
docker compose down        # 停止（PostgreSQLのデータは残る）
docker compose up --build  # 再ビルドして起動
docker compose down -v     # 停止し、PostgreSQLのデータ（Volume）も削除
```

登録したLink / Tag / Collectionは`docker compose down` → `docker compose up`をまたいで保持される（PostgreSQLのデータはDocker Volumeで永続化している）。完全に消したい場合のみ`-v`を付ける。

### Chrome拡張との併用

Backendは変わらず`http://localhost:8080`で待ち受けるため、Docker起動時もChrome拡張は追加設定なしでそのまま動作する。

## インストール方法（Dockerを使わない場合）

日常的にコードを変更しながら開発する場合は、こちらの直接起動（Frontendはホットリロード付き）が向いている。

### 必要なもの

* Go 1.26以上
* Node.js 22以上
* PostgreSQL（ローカルで起動していること）

### 1. リポジトリを取得

```bash
git clone https://github.com/matsu0122-png/linkvault.git
cd linkvault
```

### 2. データベースを準備

PostgreSQLに`linkvault`という名前でデータベースを作成し、マイグレーションを順に適用する。

```bash
createdb linkvault
for f in backend/migrations/*.sql; do
  psql -d linkvault -f "$f"
done
```

### 3. バックエンドを起動

```bash
cd backend
DB_HOST=localhost DB_PORT=5432 DB_USER=<あなたのDBユーザー名> DB_PASSWORD=<あなたのDBパスワード> DB_NAME=linkvault DB_SSLMODE=disable go run .
```

`http://localhost:8080`で起動する。環境変数を省略した場合のデフォルト値は以下の通り（`backend/config/config.go`参照）。

| 環境変数 | デフォルト値 | 説明 |
| --- | --- | --- |
| `DB_HOST` | `localhost` | PostgreSQLのホスト |
| `DB_PORT` | `5432` | PostgreSQLのポート |
| `DB_USER` | `postgres` | DBユーザー名 |
| `DB_PASSWORD` | （空） | DBパスワード |
| `DB_NAME` | `linkvault` | DB名 |
| `DB_SSLMODE` | `disable` | PostgreSQLのSSLモード |
| `CORS_ALLOWED_ORIGIN` | `http://localhost:5173` | 許可するオリジン（カンマ区切りで複数指定可。Chrome拡張利用時は拡張IDも追加する） |

### 4. フロントエンドを起動

別ターミナルで以下を実行する。

```bash
cd frontend
npm install
npm run dev
```

`http://localhost:5173`をブラウザで開く。

## Roadmap

* [x] リンクのCRUD（登録・一覧・編集・削除）
* [x] タグの設定
* [x] タグによる絞り込み
* [x] キーワード検索
* [x] URLからのタイトル自動取得
* [x] メタデータ自動取得（OGP説明文・favicon）
* [x] 複数URLの一括登録
* [x] 保存したURLの生存確認・リンク切れ検知
* [ ] AIによる要約・自動タグ生成（対応方針は検討中）
* [x] Chrome拡張からの保存
* [x] ページネーション
* [x] Collectionによるリンクのグループ管理
* [ ] より高度な全文検索

## プロジェクトについて

### 名前の由来

**LinkVault（リンク・ヴォルト）**

「Link」と「Vault」を組み合わせた名前です。

Vaultには「金庫」「保管庫」という意味があります。

Web上で見つけた有用なリンクを大切に保管し、必要になったときに取り出せる場所という意味を込めています。

### 開発者

matsuyamashin

### ライセンス

[MIT License](LICENSE)

### バージョン

v1.0.0

## 開発状況

v1.0として一通りの機能・テスト・ドキュメントを整備済み。今後の予定は上記Roadmap、および[`.github/assets/spec.md`](.github/assets/spec.md)の「将来的な拡張」を参照。
