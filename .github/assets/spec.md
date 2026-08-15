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

`service`パッケージは呼び出し先（`linkRepository`・`collectionRepository`・`metadataFetcher`）のインターフェースを自パッケージ内で定義し、`repository`・`fetcher`パッケージの具象型を受け取る（Goの「利用側でインターフェースを定義する」慣習に従う）。`LinkService`と`CollectionService`はどちらも`service`パッケージ内の別ファイルとして実装し、Collection用に新しいパッケージを分けることはしていない（既存の`handler`/`repository`/`model`もentityごとにファイルを分けるだけで、レイヤーパッケージ自体は1つのまま）。

### エラーハンドリングの方針

`handler`はエラーを大きく3種類に分けて扱う。

1. **`errors.Is`で判定できる既知のエラー**（`service.ErrNotFound`・`service.ErrDuplicateName`等）→ それぞれ意味に応じた明確なステータス（404・409等）
2. **`*service.ValidationError`**（`errors.As`で判定）→ ユーザー入力が原因の失敗として`400 Bad Request`、メッセージをそのままクライアントへ返す（例:「url is required」）
3. **上記のどちらでもないエラー**（DB接続断など想定外の失敗）→ `500 Internal Server Error`を返し、詳細はクライアントに公開せずサーバーログにのみ記録する

この分類は`handler.writeServiceError`という共通ヘルパーに集約している。`service`層のバリデーション関数（`CreateLink`の`url is required`、`validateCollectionName`等）は`errors.New`ではなく`*service.ValidationError`を返すことで、この分類がhandler側で機械的に行えるようにしている。以前は「既知のエラー以外はすべて400・エラー内容をそのまま返す」という実装になっており、内部エラーの誤ったステータスコードとエラー詳細の露出という問題があったため、v1.0でこの方式に修正した。

### 実行構成

LinkVaultは「ローカルに直接プロセスを起動する」「Docker Composeでコンテナとして起動する」の2通りの実行方法を持つが、アプリケーションコード（Go/TypeScript）はどちらの方法でも同一で、実行方法を意識した分岐は一切含まない。すべて環境変数経由の設定注入（`config.Load()` / `import.meta.env.VITE_API_BASE_URL`）で完結しており、これはDocker化以前から成立していた設計である（具体的なコマンドはREADME.mdの該当節を参照。ここでは重複を避けるため仕組みのみ記す）。

* **backend↔db間の接続先**: Docker Compose環境では`DB_HOST`にPostgreSQLコンテナのservice name（`db`）を渡す。これはコンテナ間の内部ネットワーク通信であり、ブラウザやChrome拡張からは見えない
* **frontend→backend間の接続先**（`VITE_API_BASE_URL`）: **ブラウザから見えるURL**（例: `http://localhost:8080`）を渡す。Viteのビルド時に静的ファイルへ埋め込まれる値のため、コンテナ起動後に変更しても反映されない（再ビルドが必要）。`backend`のservice name（`http://backend:8080`）を渡すのは誤り（ブラウザはDocker内部ネットワークを解決できない）
* **CORS**: `CORS_ALLOWED_ORIGIN`はDocker化の有無に関わらず`http://localhost:5173`のまま変わらない。ブラウザが実際にアクセスするオリジンはDocker化前後で変化しないため
* **migration適用**: Docker環境では、PostgreSQL公式イメージの`/docker-entrypoint-initdb.d/`にマウントした`backend/migrations/*.sql`を、データディレクトリが空の初回起動時にファイル名順（`0001_`〜）で自動実行させている。CI（GitHub Actions）で使っている「`psql -f`をファイル名順に適用する」というmigration方式そのものは変更していない。Docker用の別migrationツールは導入していない

## データ設計

PostgreSQLで以下の5テーブルを管理する。

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
| status      | TEXT        | 生存確認の結果。`unknown` / `ok` / `broken`のいずれか（`CHECK`制約、`NOT NULL DEFAULT 'unknown'`） |
| checked_at  | TIMESTAMPTZ | 最終チェック日時（未チェックならNULL）        |
| created_at  | TIMESTAMP   | 登録日時                  |
| updated_at  | TIMESTAMP   | 最終更新日時                |

`description` / `image_url` / `favicon_url` / `status` / `checked_at`はユーザーが直接編集する手段を持たない（`PUT`のリクエストボディにも含まれない）。前者3つは`POST`時の自動取得、後者2つは`POST /api/links/check`でのみ設定され、以降は`Update`で上書きされることなく保持される。

`checked_at`のみ他の日時カラムと異なり`TIMESTAMPTZ`（タイムゾーン付き）にしている。`created_at`/`updated_at`は初期マイグレーションから`TIMESTAMP`（タイムゾーン無し）のままだが、これはアプリケーションが常にサーバーのローカル時刻で書き込む前提で運用されており、これまで実害が出ていないため据え置いている。`checked_at`は実装時に`TIMESTAMP`で試したところ、書き込み時と読み出し時でオフセットがずれる不具合が実際に発生したため、素直に`TIMESTAMPTZ`を採用した。

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

### collections

Tagとは別に、リンクを「どのグループにまとめるか」を表すCollectionを管理する（Tagは「どんな属性を持つか」を表す。両者の詳細な役割の違いは後述「## Collection」参照）。

| 項目          | 型         | 説明                              |
| ----------- | --------- | ------------------------------- |
| id          | BIGSERIAL | 主キー                             |
| name        | TEXT      | Collection名（`UNIQUE`制約）         |
| description | TEXT      | Collectionの説明（`NOT NULL DEFAULT ''`） |
| parent_id   | BIGINT    | 親Collectionの`id`への自己参照外部キー（`NULL`可、`ON DELETE SET NULL`） |
| created_at  | TIMESTAMP | 登録日時                            |
| updated_at  | TIMESTAMP | 最終更新日時                          |

`name`を`UNIQUE`制約にしているのは`tags.name`と同じ理由（後述「## Collection」の「Collection名の重複」参照）。

`parent_id`はCollectionの入れ子構造（Collectionの中にさらにCollectionを作る）を表す自己参照外部キー。`NULL`は最上位のCollectionを意味する。`ON DELETE SET NULL`により、親Collectionを削除しても子Collectionは削除されず、`parent_id`が`NULL`になって最上位へ昇格する（詳細は後述「## Collection」の「入れ子構造」参照）。

### collection_links

`links`と`collections`のN:N関係を表す中間テーブル。`link_tags`と同じ構造を採用する。

| 項目            | 型      | 説明                                       |
| ------------- | ------ | ---------------------------------------- |
| collection_id | BIGINT | `collections.id`への外部キー（`ON DELETE CASCADE`） |
| link_id       | BIGINT | `links.id`への外部キー（`ON DELETE CASCADE`）      |

主キーは`(collection_id, link_id)`の複合キーとし、同一リンクの同一Collectionへの重複登録をDBレベルで防止する。`collection_id`側のCASCADEにより、Collectionを削除しても`links`本体は削除されず、`collection_links`の関連行のみが削除される（詳細は「## Collection」参照）。

## API設計

フロントエンドとバックエンドはJSON形式のHTTP APIで通信する。

### リンク一覧の取得

```text
GET /api/links
```

クエリパラメータ（いずれも省略可）：

| パラメータ   | 説明                                                     |
| ------- | ------------------------------------------------------ |
| `q`     | `url` / `title` / `memo`への部分一致（大文字小文字を区別しない）            |
| `tag`   | タグ名の完全一致による絞り込み                                         |
| `page`  | ページ番号（1始まり）。デフォルト`1`                                    |
| `limit` | 1ページあたりの件数。デフォルト`20`、最大`100`                             |
| `sort`  | 並び順。`created_at_desc`（デフォルト） / `created_at_asc` / `updated_at_desc` / `updated_at_asc` / `title_asc` / `title_desc` |
| `collection` | Collection IDによる絞り込み。指定したCollectionに所属するリンクのみ返す |

`q` / `tag` / `collection`による絞り込みと`page`/`limit`/`sort`は自由に組み合わせられる（例: `GET /api/links?q=golang&tag=Go&collection=3&page=2&limit=20&sort=title_asc`）。`page` / `limit` / `sort` / `collection`は未指定・空・数値として不正・許可された値の範囲外のいずれであっても`400`にはならず、それぞれのデフォルト値（`collection`は「絞り込み無し」）へフォールバックする（`limit`が最大値を超える場合は最大値へクランプする）。これは`q` / `tag`が元々どんな値でも受け付ける（バリデーション無し）方針と揃えたもので、ページネーション・並び替え・Collection絞り込みは「表示上の調整」であり、リクエスト自体を失敗させる理由にはしない、という設計判断による。存在しないCollection IDを指定した場合はエラーにはならず、単に0件が返る（`tag`に存在しない値を指定した場合と同じ扱い）。

出力例：

```json
{
  "links": [
    {
      "id": 1,
      "url": "https://go.dev/doc/",
      "title": "Go Documentation",
      "memo": "Goの公式ドキュメント",
      "tags": ["Go", "backend"],
      "description": "Go言語の公式ドキュメント",
      "image_url": "https://go.dev/images/og-image.png",
      "favicon_url": "https://go.dev/favicon.ico",
      "status": "ok",
      "checked_at": "2026-08-14T10:00:00Z",
      "created_at": "2026-08-14T09:30:00Z",
      "updated_at": "2026-08-14T09:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 83,
    "totalPages": 5
  }
}
```

`pagination.total`は`q` / `tag`による絞り込み適用後の件数（現在のページの件数ではなく、全ページ合計の件数）。`totalPages`は`ceil(total / limit)`（`total`が0の場合は`0`）。

**並び順の安定性**: `sort`が指定するカラムだけでは同値が発生しうる（例: 同時刻に一括登録された複数リンクの`created_at`が同じ、あるいは複数リンクが同じタイトル）ため、必ず`id`を第2キーとして付加する（`created_at_desc`なら`ORDER BY created_at DESC, id DESC`）。これにより、ページを移動してもリンクの重複・欠落が起きない。

**SQLインジェクション対策**: `ORDER BY`句はプレースホルダで値を渡せない（カラム名・方向を文字列結合で組み立てる必要がある）ため、`sort`の生の文字列を直接SQLへ混ぜることはしない。`service`層で許可された6値のいずれかにマッチするかを検証し（不一致ならデフォルトへフォールバック）、`repository`層でも改めて許可値ごとに固定のORDER BY文字列を返すswitch文を経由させる二重の防御を行っている。

`GET /api/links`が返す一覧の総件数・ページ内容は、`WHERE`句（`q` / `tag` / `collection`条件）をGoの文字列定数として1箇所にまとめ、件数取得（`COUNT(*)`）と一覧取得の両方のSQLで共有することで、フィルタ条件のズレによる件数とページ内容の不整合を防いでいる。

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
  "status": "unknown",
  "checked_at": null,
  "created_at": "2026-08-14T09:30:00Z",
  "updated_at": "2026-08-14T09:30:00Z"
}
```

新規作成したリンクは常に`status: "unknown"` / `checked_at: null`で始まる（`POST`時に生存確認は行わない）。`url`は必須。`url`が空の場合は`400 Bad Request`を返す。`title` / `description` / `image_url` / `favicon_url`の扱いは次節「リンク登録仕様」を参照。リクエストボディに`description` / `image_url` / `favicon_url`を含めても無視される（これらはユーザー入力を受け付けない）。

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

### リンクの一括登録

```text
POST /api/links/bulk
```

複数のURLをまとめて登録する。`title`/`memo`は指定できず、常にメタデータ自動取得のみで決まる（単体登録の`title`空文字時と同じ扱い）。`tags`は全URLに共通で適用される。

入力例：

```json
{
  "urls": ["https://go.dev", "https://react.dev", ""],
  "tags": ["reading-list"]
}
```

出力例：

```json
{
  "created": [
    { "id": 10, "url": "https://go.dev", "title": "The Go Programming Language", "...": "..." },
    { "id": 11, "url": "https://react.dev", "title": "React", "...": "..." }
  ],
  "failed": [
    { "url": "", "error": "url is required" }
  ]
}
```

`urls`が空、または`MaxBulkURLs`（50件）を超える場合は`400 Bad Request`を返す。それ以外は常に`200 OK`で、`created`と`failed`の内訳をレスポンスボディに含める（1件の失敗が他のURLの登録を妨げない）。`created` / `failed`はいずれも`urls`の入力順を保持する。

`failed`に載るのは`url`が空文字などの明確な不正入力のみで、メタデータ取得の失敗は単体登録と同様に非致命的に扱われる（`title`等が空のまま`created`に含まれる）。

サーバー内部では、各URLの処理（メタデータ取得＋DB保存）を`service.LinkService.CreateLink`にそのまま委譲し、goroutine + セマフォ（チャネル、上限`maxConcurrentFetches`＝5）で並行実行する。並行数を制限するのは、自サーバー・取得先サイト双方への負荷を抑えるため。

### リンクの生存確認

```text
POST /api/links/check
```

保存されている**全リンク**を対象に、現在アクセス可能かどうかを再チェックする。リクエストボディは無し。定期実行するバックグラウンドジョブは無く、このAPIを呼ぶことでのみ実行される（手動トリガーのみ）。フロントエンドは「リンク切れをチェック」ボタンから呼び出す。

出力はチェック後の全リンク一覧を配列でそのまま返す（`{"links": [...], "pagination": {...}}`のようなページネーションのenvelopeは持たない）。「全リンクの状態を一括更新する」操作であり、ページや並び替えといった表示上の概念とは無関係なため、`GET /api/links`とはレスポンス形式が異なる。

チェック対象のリンクが多い場合でも、`POST /api/links/bulk`と同じgoroutine + セマフォ（上限`maxConcurrentFetches`＝5）で並行にチェックする。1件のチェック失敗（タイムアウト・接続エラー等）が他のリンクのチェックを妨げない。

各リンクについて`fetcher.CheckAlive(url string) (bool, error)`を呼び、以下のルールで`status`を更新する。

* 2xx（リダイレクト追従後）が返れば`status = "ok"`
* それ以外（4xx/5xx・接続失敗・タイムアウト・SSRF対策によるアクセス拒否など）はすべて`status = "broken"`として扱う。ユーザーにとって重要なのは「今アクセスできるかどうか」の一点のみであり、失敗理由の細分化はAPI仕様上行わない（サーバーログには`log.Printf`で理由を記録する）
* `checked_at`は実行時刻（サーバー時刻）で更新する

`CheckAlive`は`FetchMetadata`と同じ`http.Client`（SSRF対策・タイムアウト設定込み）を使い、GETリクエストのステータスコードのみを見る（ボディはHTMLパースせず読み捨てる）。

## Collection

### 目的とTagとの違い

Tagが「リンクがどんな属性を持つか」を表すのに対し、Collectionは「リンクをどのグループにまとめるか」を表す。両者は独立した機能であり、一方が他方を置き換えるものではない。

| | Tag | Collection |
| --- | --- | --- |
| 役割 | リンクの属性ラベル | リンクのグループ分け |
| Linkとの関係 | N:N（`link_tags`） | N:N（`collection_links`） |
| 独自メタ情報 | 名前のみ | 名前・説明文 |
| 入力UI | フリーテキスト＋チップ入力（`TagInput`） | 既存Collectionからのチェックボックス選択（`CollectionSelect`） |

1つのLinkは複数のCollectionへ同時に所属できる（例: 「Go学習」と「お気に入り」の両方）。

### 入れ子構造

Collectionは`parent_id`により、Collectionの中にさらにCollectionを作る入れ子構造を持てる（例: 「お気に入り」の中に「新作アニメ」を作る）。深さの制限は設けていない。

**親の指定は作成時のみ**: `parent_id`は`POST /api/collections`でのみ指定できる。作成後に別の場所へ移動する（親を変更する）APIは存在しない。この制約により、「あるCollectionを自分自身の子孫の下へ移動して循環参照を作ってしまう」というケースが構造的に発生しない（新しく作られたCollectionにはまだ子孫が存在しないため、循環のしようがない）。将来的に移動機能を追加する場合は、移動先が自分自身の子孫でないかを確認する循環検出ロジックが別途必要になる。

**親の削除**: 親Collectionを削除しても子Collectionは削除されない。`parent_id`カラムの`ON DELETE SET NULL`により、子は自動的に最上位（`parent_id = NULL`）へ昇格する。子Collectionに紐づくLink自体も、既存の「Collection削除時の挙動」と同様に一切削除されない。

**Link一覧との関係**: `GET /api/links?collection=:id`は、そのCollectionに**直接**紐付けられたLinkのみを返す。子Collectionに属するLinkは含まない（再帰的な集計はしない）。`link_count`も同様に直属のLinkの件数のみで、子孫の件数は合算しない。子Collectionの中身を見るには、サイドバーでその子Collectionへ直接遷移する。これはSQLをシンプルに保つための設計判断であり、「親を開けば子の中身も全部見える」という体験は現時点では提供しない。

### Collectionの一覧取得

```text
GET /api/collections
```

クエリパラメータ（省略可）：

| パラメータ | 説明 |
| --- | --- |
| `link_id` | 指定した場合、そのLinkが所属するCollectionのみ返す。省略時は全件返す |

`name`の昇順で返す。出力は配列（ページネーションは行わない。個人利用規模でCollection数が多くなることは想定していないため）。

出力例：

```json
[
  {
    "id": 1,
    "name": "Go学習",
    "description": "Go関連の学習資料",
    "parent_id": null,
    "link_count": 12,
    "created_at": "2026-08-15T09:00:00Z",
    "updated_at": "2026-08-15T09:00:00Z"
  },
  {
    "id": 2,
    "name": "サブスク",
    "description": "",
    "parent_id": 1,
    "link_count": 3,
    "created_at": "2026-08-15T09:05:00Z",
    "updated_at": "2026-08-15T09:05:00Z"
  }
]
```

`link_count`は`collection_links`をCOUNTした値で、一覧取得のたびに動的に計算する（非正規化して`collections`テーブルへ保持することはしていない）。子孫Collectionの件数は含まない（前述「入れ子構造」参照）。

レスポンスは常にフラットな配列で、木構造への組み立て（`parent_id`を辿って親子関係を復元する処理）はフロントエンド側（`features/collections/tree.ts`）で行う。バックエンドは`parent_id`を返すだけで、階層を意識したレスポンス形式は持たない。

`link_id`パラメータは、Link編集フォームが「このLinkは現在どのCollectionに所属しているか」を知るために使う。Linkモデル自体には所属Collectionの情報を持たせておらず（`GET /api/links`のレスポンスに`collection_ids`のような欄はない）、必要なときにこのパラメータで都度取得する設計にしている。既存の`Link`のSQL・モデル・テストに一切手を加えずに済む、低リスクな設計を優先した。

### Collectionの登録

```text
POST /api/collections
```

入力例（最上位に作成する場合）：

```json
{ "name": "Go学習", "description": "Go関連の学習資料" }
```

入力例（既存Collectionの子として作成する場合）：

```json
{ "name": "サブスク", "description": "", "parent_id": 1 }
```

`name`は前後の空白を除去したうえで必須（空文字・空白のみは`400`）、100文字を超える場合も`400`。`name`が既存のCollectionと重複する場合は`409 Conflict`を返す（`collections.name`の`UNIQUE`制約違反を検知）。`parent_id`は省略可（省略・`null`なら最上位のCollectionとして作成）。指定した`parent_id`が存在しない場合は`404 Not Found`を返す。成功時は`201 Created`で作成されたCollection（`link_count: 0`）を返す。

### Collectionの詳細取得

```text
GET /api/collections/:id
```

指定されたIDのCollectionを返す（`GET /api/collections`の1件分と同じ形式）。対象が存在しない場合は`404 Not Found`。所属するLinkの一覧はこのAPIには含まれない。`GET /api/links?collection=:id`で別途取得する（後述）。

### Collectionの更新

```text
PUT /api/collections/:id
```

リクエストボディは`POST`と同じ形式（`name` / `description`）。バリデーション・重複チェックも`POST`と同様。対象が存在しない場合は`404 Not Found`。`collection_links`（所属Link）はこのAPIでは変更されない。

### Collectionの削除

```text
DELETE /api/collections/:id
```

指定されたIDのCollectionを削除する。成功時は`204 No Content`。対象が存在しない場合は`404 Not Found`。

**削除してもLink本体は削除されない。** `collection_links`の該当行はCASCADEで自動削除されるが、`links`テーブルの行はそのまま残る。他のCollectionへの所属や「すべてのリンク」からは引き続き参照できる。

**子Collectionも削除されない。** 削除したCollectionに子Collectionがあった場合、子は削除されず最上位（`parent_id = NULL`）へ昇格する（前述「入れ子構造」参照）。

### CollectionへのLink追加

```text
POST /api/collections/:id/links
```

入力例：

```json
{ "link_id": 5 }
```

成功時は`204 No Content`。指定したCollectionが存在しない場合は`404 Not Found`（`Collection not found`）。指定した`link_id`が存在しない場合も`404 Not Found`（`Link not found`）。

**既に追加済みのLinkを再度追加した場合**は、エラーにはせず成功（`204`）として扱う（`collection_links`への`INSERT ... ON CONFLICT (collection_id, link_id) DO NOTHING`により、状態は変化しないがリクエスト自体は成功する）。「既にそのCollectionに入っている」という状態を実現しようとした操作として扱い、冪等に成功させる設計とした。

### CollectionからのLink削除

```text
DELETE /api/collections/:id/links/:linkId
```

指定したCollectionから指定したLinkの関連付けのみを削除する（Link本体は削除されない）。成功時は`204 No Content`。指定した組み合わせが元々紐付いていない場合は`404 Not Found`。

### Link一覧・Link編集フォームとの連携

- `GET /api/links`の`collection`パラメータ（前述）により、「特定のCollectionに所属するLinkだけを検索・タグ絞り込み・ページネーション・並び替え付きで一覧取得する」機能を、新しいエンドポイントを追加せずに実現している。Collection詳細画面はこの仕組みの上に成り立っており、`GET /api/collections/:id`（メタ情報）と`GET /api/links?collection=:id`（所属Link一覧）を組み合わせて表示する
- Link作成・編集フォームのCollection選択は、`POST /api/collections`ではなく`POST /api/collections/:id/links` / `DELETE /api/collections/:id/links/:linkId`を通じて行う。Linkの作成・更新（`POST` / `PUT /api/links`）のリクエストボディに`collection_ids`のような欄は存在しない。フロントエンドはLink保存後に、選択されたCollectionとの差分（追加分・削除分）だけを計算し、上記2つのAPIを必要な回数だけ呼び出す

## クライアントとCORS

バックエンドAPIを叩くクライアントは2つある。

* `frontend/`: React + TypeScript製のWebアプリ（デフォルト`http://localhost:5173`で動作）
* `extension/`: Chrome拡張機能（後述）。オリジンは`chrome-extension://<拡張ID>`で、拡張IDは「パッケージ化されていない拡張機能」として読み込むたびに環境依存で変わる

このため`CORS_ALLOWED_ORIGIN`環境変数はカンマ区切りで複数オリジンを受け付ける（`config.parseOrigins`）。`main.go`の`withCORS`ミドルウェアは、リクエストの`Origin`ヘッダーが許可リストに含まれる場合のみ、そのオリジンをそのまま`Access-Control-Allow-Origin`として返す（固定値を常に返す方式ではない）。リストに無いオリジンにはCORSヘッダーを一切付与せず、ブラウザ側でブロックさせる。

### Chrome拡張（`extension/`）

Manifest V3。ビルドステップを持たないプレーンなHTML/CSS/JSで構成する（`manifest.json` / `popup.html` / `popup.css` / `popup.js`）。

* 権限は`activeTab`のみ。常時ブラウジング履歴やタブ一覧へアクセスする権限は要求せず、ユーザーが拡張アイコンをクリックした瞬間のアクティブタブにのみ、その場でアクセスする
* ポップアップを開くと`chrome.tabs.query`で現在のタブの`url`/`title`を取得し、タイトル欄に初期値として表示する（編集可能）
* 「保存」ボタンで`POST /api/links`をブラウザから直接呼び出す（バックエンドの`API_BASE_URL`は`http://localhost:8080`に固定。設定UIは現時点では持たない）
* タグ入力はWebアプリのようなチップ式ではなく、カンマ区切りのテキスト入力（ビルド不要の軽量実装を優先したための簡略化）
* `title`を空のまま保存した場合は、既存の`POST /api/links`の仕様通りサーバー側でメタデータ自動取得が行われる

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

生存確認（`POST /api/links/check`）は手動トリガーのみで、サーバー起動中に自動で定期実行する仕組みは持たない。バックグラウンドでの定期チェック（`time.Ticker`によるスケジューリングなど）は将来検討課題として残っている。

Chrome拡張（`extension/`）は現時点でタグ入力欄がカンマ区切りテキストのみで、Webアプリのようなチップ式UIやAPIベースURLの設定画面は持たない。拡張機能ストアへの公開も行っていない（ローカルでの「パッケージ化されていない拡張機能」読み込みのみを想定）。

その他、README.mdのRoadmapに記載の以下も未実装である。

* AIによる要約・自動タグ生成
* より高度な全文検索（現在は`ILIKE`による部分一致のみ。ページネーション自体は実装済みで、全文検索が高度化されても`page`/`limit`/`sort`の仕組みはそのまま使える設計になっている）

これらは現時点で設計・実装を行っていないため、対応するフィールドや抽象化を先回りして追加することはしない。

Collection機能についても、以下は今回実装していない（将来検討）。

* Chrome拡張から保存時にCollectionを指定する機能（現状`extension/`はCollection非対応）
* Collectionのアイコン・色分け
* Collectionの並び替え（現状は`name`昇順固定）
* Collectionのお気に入り登録
* Collectionの共有・公開URL（例:「Go初心者におすすめのサイト集」を外部に公開する）
* Collection単位でのリンクのエクスポート
* サイドバーでのTag一覧表示（現状はLink行のタグピルをクリックする既存の絞り込み動線のみ。Tag一覧を返すAPIも現時点では存在しない）

これらを見越した抽象化（例: Collectionへの`icon`カラムや`sort_order`カラムの先行追加）は行っていない。
