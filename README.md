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
* 複数URLのまとめて登録
* 登録したリンクの一覧表示
* リンクの編集・削除
* タグによる整理・絞り込み
* メモの保存
* キーワードによる検索

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

## インストール方法

現在開発中のため、インストール方法は未定です。

開発の進行に合わせて追記します。

## 使用技術

### フロントエンド

* React
* TypeScript
* Vite
* Tailwind CSS

### バックエンド

* Go

### データベース

* PostgreSQL

### CI / テスト

* GitHub Actions
* Coveralls

## Roadmap

* [x] リンクのCRUD（登録・一覧・編集・削除）
* [x] タグの設定
* [x] タグによる絞り込み
* [x] キーワード検索
* [x] URLからのタイトル自動取得
* [x] メタデータ自動取得（OGP説明文・favicon）
* [x] 複数URLの一括登録
* [ ] 保存したURLの生存確認・リンク切れ検知
* [ ] AIによる要約・自動タグ生成
* [ ] Chrome拡張からの保存
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

未定

### バージョン

未定

### バージョン履歴

開発開始後に記載します。

## 開発状況

現在開発中です。
