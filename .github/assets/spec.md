# LinkVault Specification

このドキュメントは、LinkVaultの技術的な仕様を記録するものである。README.mdがプロジェクト全体の紹介を目的とするのに対し、本ドキュメントは開発者がAPI・データ設計・セキュリティ仕様などを正確に把握するためのものとする。

内容は実装と一致させる。将来的な拡張案は「将来的な拡張」節に明記し、現在の仕様と混同しない。

## アーキテクチャ

バックエンド（Go）は以下の層で構成する。

```text
main.go        起動処理の組み立てのみ（設定読み込み・DB接続・ルーティング初期化）
handler/       HTTPリクエスト/レスポンスの変換
service/       バリデーション・ビジネスロジック
repository/    SQL・DBアクセス
model/         データ構造
fetcher/       外部WebページからのHTMLメタデータ取得
config/        環境変数の読み込み
database/      DB接続
```

`service`パッケージは呼び出し先（`linkRepository`・`metadataFetcher`）のインターフェースを自パッケージ内で定義し、`repository`・`fetcher`パッケージの具象型を受け取る（Goの「利用側でインターフェースを定義する」慣習に従う）。

## データ設計

PostgreSQLで以下の3テーブルを管理する。

### links

| 項目         | 型           | 説明                    |
| ---------- | ----------- | --------------------- |
| id          | BIGSERIAL   | 主キー                   |
| url         | TEXT        | 保存するWebページのURL（必須）    |
| title       | TEXT        | リンクのタイトル（空文字を許容）      |
| memo        | TEXT        | ユーザーが自由に記入するメモ        |
| description | TEXT        | ページの説明文（自動取得、`NOT NULL DEFAULT ''`） |
| image_url   | TEXT        | OGP画像のURL（自動取得、`NOT NULL DEFAULT ''`） |
| favicon_url | TEXT        | faviconのURL（自動取得、`NOT NULL DEFAULT ''`） |
| created_at  | TIMESTAMP   | 登録日時                  |
| updated_at  | TIMESTAMP   | 最終更新日時                |

`description` / `image_url` / `favicon_url`はユーザーが直接編集する手段を持たない（`PUT`のリクエストボディにも含まれない）。`POST`時の自動取得でのみ設定され、以降は`Update`で上書きされることなく保持される。

### tags

| 項目   | 型         | 説明               |
| ---- | --------- | ---------------- |
| id   | BIGSERIAL | 主キー               |
| name | TEXT      | タグ名（UNIQUE制約）     |

### link_tags

`links`と`tags`のN:N関係を表す中間テーブル。

| 項目      | 型      | 説明                                   |
| ------- | ------ | ------------------------------------ |
| link_id | BIGINT | `links.id`への外部キー（`ON DELETE CASCADE`） |
| tag_id  | BIGINT | `tags.id`への外部キー（`ON DELETE CASCADE`）  |

主キーは`(link_id, tag_id)`の複合キーとし、同一リンクへの同一タグの重複紐付けを防ぐ。

タグ名はアプリケーション側（`service.normalizeTags`）で前後の空白除去・重複除去を行ってから保存する。

## API設計

フロントエンドとバックエンドはJSON形式のHTTP APIで通信する。

### リンク一覧の取得

```text
GET /api/links
```

クエリパラメータ（いずれも省略可）：

| パラメータ | 説明                                          |
| ----- | ------------------------------------------- |
| `q`   | `url` / `title` / `memo`への部分一致（大文字小文字を区別しない） |
| `tag` | タグ名の完全一致による絞り込み                             |

出力例：

```json
[
  {
    "id": 1,
    "url": "https://go.dev/doc/",
    "title": "Go Documentation",
    "memo": "Goの公式ドキュメント",
    "tags": ["Go", "backend"],
    "description": "Go言語の公式ドキュメント",
    "image_url": "https://go.dev/images/og-image.png",
    "favicon_url": "https://go.dev/favicon.ico",
    "created_at": "2026-08-14T09:30:00Z",
    "updated_at": "2026-08-14T09:30:00Z"
  }
]
```

### リンクの登録

```text
POST /api/links
```

入力例（タイトルを自動取得させる場合）：

```json
{
  "url": "https://example.com",
  "title": "",
  "memo": "",
  "tags": ["Go", "web"]
}
```

出力例：

```json
{
  "id": 1,
  "url": "https://example.com",
  "title": "Example Domain",
  "memo": "",
  "tags": ["Go", "web"],
  "description": "This domain is for use in illustrative examples.",
  "image_url": "",
  "favicon_url": "",
  "created_at": "2026-08-14T09:30:00Z",
  "updated_at": "2026-08-14T09:30:00Z"
}
```

`url`は必須。`url`が空の場合は`400 Bad Request`を返す。`title` / `description` / `image_url` / `favicon_url`の扱いは次節「リンク登録仕様」を参照。リクエストボディに`description` / `image_url` / `favicon_url`を含めても無視される（これらはユーザー入力を受け付けない）。

### リンクの更新

```text
PUT /api/links/:id
```

指定されたIDのリンクを更新する。リクエストボディは`POST`と同じ形式（`url` / `title` / `memo` / `tags`）。`tags`は送信された内容で完全に置き換えられる（差分更新ではない）。対象が存在しない場合は`404 Not Found`。

`PUT`では`POST`と異なり、メタデータの自動取得は行わない。`title`が空でも再取得はせず、`description` / `image_url` / `favicon_url`は登録時に取得した値がそのまま保持される（`UPDATE`文がこれらのカラムに触れないため）。

### リンクの削除

```text
DELETE /api/links/:id
```

指定されたIDのリンクを削除する。成功時は`204 No Content`。対象が存在しない場合は`404 Not Found`。`links`の削除に伴い、`link_tags`の該当行はCASCADEで自動削除される（`tags`自体は他のリンクから参照されている可能性があるため削除しない）。

## リンク登録仕様（メタデータ自動取得）

`POST /api/links`のみに適用される仕様。

* メタデータ取得は`title`の有無に関わらず**常に**実行する（`url`へ1回だけHTTP GETする）
* `title`が指定されている場合 → ユーザーが指定した`title`をそのまま使用する（取得結果のtitleは破棄する）
* `title`が空文字の場合 → 取得できたtitleを使用する（取得できなければ空文字のまま）
* `description` / `image_url` / `favicon_url`は、`title`の指定有無に関わらず常に取得結果を使用する（ユーザーが入力する手段がないため）

**重要**: メタデータ取得に失敗しても`POST`自体は失敗させない。取得できなかったフィールドは空文字のまま保存する。外部Webサイトの状態（応答なし・エラー・SSRF対象など）によって、ユーザーのリンク保存操作が失敗する体験を避けるための設計判断である。フロントエンドは`title`が空の場合は`url`を、`memo`が空の場合は`description`を代わりに表示する。

## メタデータ取得仕様

`fetcher`パッケージ（`FetchMetadata(url string) (model.Metadata, error)`）が以下の流れで処理する。

```text
URL
  ↓
URL parse・スキーム検証（http/https以外は拒否）
  ↓
安全性チェック（SSRF対策、後述）
  ↓
HTTP GETリクエスト（タイムアウト付き）
  ↓
HTTPレスポンス受信
  ↓
ステータスコード検証（2xx以外は失敗として扱う）
  ↓
レスポンスボディをHTMLとしてストリーム解析（golang.org/x/net/html、</head>まで）
  ↓
<title> / meta[og:description] / meta[name=description] / meta[og:image] / link[rel=icon]を抽出
  ↓
service層がLinkのTitle（未指定時のみ）・Description・ImageURL・FaviconURLへ反映
  ↓
repository層がPostgreSQLへ保存
```

抽出の優先順位・詳細：

* description: `og:description`を優先し、無ければ`<meta name="description">`
* image: `og:image`のみ（フォールバックなし）
* favicon: `<link rel="icon">`または`<link rel="shortcut icon">`のみ（`/favicon.ico`への当て推量アクセスは行わない）
* image・faviconの`href`/`content`が相対URLの場合、レスポンスの最終URL（リダイレクト後）を基準に絶対URLへ解決する
* 解決した絶対URLのスキームが`http`/`https`以外（`data:` / `javascript:`など）の場合は採用せず空文字のままにする
* `</head>`が出現した時点で走査を打ち切る（対象要素は`<head>`内にのみ存在するため）

## エラー仕様

`fetcher.FetchMetadata`が失敗するケースと、その結果`POST /api/links`がどう振る舞うかを以下にまとめる。

| ケース                          | `POST /api/links`のレスポンス | 保存されるtitle/description/image_url/favicon_url |
| ---------------------------- | ------------------------ | ----------- |
| URLのパース失敗                     | `201 Created`（保存は成功）     | すべて空文字（titleはユーザー指定があればそれを使用） |
| サポート外スキーム（`http`/`https`以外）   | `201 Created`             | 同上 |
| SSRF対策によるアクセス拒否               | `201 Created`             | 同上 |
| DNS解決失敗・接続失敗                  | `201 Created`             | 同上 |
| タイムアウト                        | `201 Created`             | 同上 |
| HTTP 4xx / 5xx                | `201 Created`             | 同上 |
| レスポンスHTMLに該当要素が存在しない          | `201 Created`             | 見つからなかったフィールドのみ空文字 |

すべてのケースで`POST`自体は成功として扱う（`url`が空という唯一のバリデーションエラーのみ`400`を返す）。取得失敗の詳細はサーバーログ（`log.Printf`）にのみ記録し、APIレスポンスには含めない。SSRF拒否についても他の失敗と同じ扱いとし、クライアントにネットワーク構成の手がかりを与えない。

## セキュリティ仕様（SSRF対策）

`POST /api/links`はユーザーが指定した任意のURLへサーバーが直接HTTPアクセスするため、SSRF（Server-Side Request Forgery）対策を実施している。

**実装済みの対策**:

* URLスキームを`http` / `https`のみに制限する（`file://`などを拒否）
* `net.Dialer`の`Control`フックで、実際に接続を確立する直前の生IPアドレスを検査し、以下に該当する場合は接続を拒否する
  * loopback（`127.0.0.0/8`、`::1`）
  * private（`10.0.0.0/8`、`172.16.0.0/12`、`192.168.0.0/16`など、`net.IP.IsPrivate()`が真となる範囲）
  * link-local（`169.254.0.0/16`、`fe80::/10`。クラウド環境のメタデータエンドポイント`169.254.169.254`もここに含まれる）
  * unspecified（`0.0.0.0`、`::`）
* `Control`フックはDNS解決後・接続確立前の実IPアドレスに対して呼ばれるため、検証後にDNS応答を変えるDNS rebinding攻撃に対しても有効に機能する
* リダイレクトは許可するが最大3回までとし（`http.Client.CheckRedirect`）、リダイレクト先も同じ`Transport`経由で接続するため上記の`Control`フックによる検証を毎回受ける

**実装していない対策（将来検討）**:

* IPv6のULA（`fc00::/7`）やその他の特殊アドレス範囲の網羅的な除外
* 許可ドメインのホワイトリスト方式
* レスポンスヘッダー（`Content-Type`）による事前フィルタリング

## タイムアウト

| 設定項目                      | 値       |
| ------------------------- | ------- |
| HTTPリクエスト全体のタイムアウト        | 5秒      |
| TCP接続確立のタイムアウト            | 3秒      |
| レスポンスボディの読み取り上限           | 512KB   |
| リダイレクト回数の上限               | 3回      |

**採用理由**: リンクの登録はユーザーが画面を見ながら待つ操作であるため、応答の遅い外部サイトのために体感速度を大きく損なわないことを優先した。5秒あれば大半のWebサイトはヘッダー〜`<head>`終了までを返せる一方、応答がないサイトを無期限に待つことは避けられる。`title` / `description` / `image` / `favicon`はいずれも通常HTMLの`<head>`内、先頭数KB以内に出現するため、512KBの読み取り上限で実用上ほぼ全てのケースをカバーできる。値は固定値としてコード内に定義しており、現時点では環境変数などによる外部設定は行っていない。

## 将来的な拡張

`fetcher`は現在title・description・OGP画像・faviconを取得する（`FetchMetadata`が`model.Metadata`を返す）。取得した`image_url`はAPIレスポンスに含まれるが、現時点でフロントエンドには表示していない（一覧UIはfaviconとdescription/memoのみを表示する設計）。

拡張候補（未実装）：

```text
Metadata
├── Title         （実装済み）
├── Description   （実装済み: og:description優先、フォールバックでmeta description）
├── ImageURL       （実装済み: 取得・保存のみ。UI表示は未実装）
├── FaviconURL     （実装済み）
└── OGP Title（og:title、現在はHTML<title>のみを使用）
```

その他、README.mdのRoadmapに記載の以下も未実装である。

* 保存したURLの生存確認・リンク切れの定期チェック
* 複数URLの一括登録
* AIによる要約・自動タグ生成
* Chrome拡張からの保存
* より高度な全文検索（現在は`ILIKE`による部分一致のみ）

これらは現時点で設計・実装を行っていないため、対応するフィールドや抽象化を先回りして追加することはしない。
