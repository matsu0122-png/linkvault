# LinkVault

[![build](https://github.com/matsu0122-png/linkvault/actions/workflows/build.yaml/badge.svg)](https://github.com/matsu0122-png/linkvault/actions/workflows/build.yaml)

[![Coverage Status](https://coveralls.io/repos/github/matsu0122-png/linkvault/badge.svg?branch=main)](https://coveralls.io/github/matsu0122-png/linkvault?branch=main)

> Webで見つけた大切な情報を、あとから簡単に見つけられるように。



![Status](https://img.shields.io/badge/status-development-orange)

## 概要

LinkVaultは、Web上で見つけた記事、技術ドキュメント、GitHubリポジトリなどのURLを保存・整理し、あとから簡単に探せるようにするWebアプリケーションです。

Web上で有用な情報を見つけても、時間が経つと「どこで見つけたのか」「何のために保存したのか」が分からなくなることがあります。

LinkVaultでは、URLだけでなくタイトル、タグ、メモなどの情報を一緒に保存することで、必要な情報をあとから検索・整理できるようにします。

また、保存したURLが現在もアクセス可能か確認し、リンク切れなどを検出できる機能も提供します。

## 主な機能

* URLの登録
* 登録したリンクの一覧表示
* リンクの編集・削除
* タグの設定
* メモの保存
* キーワードによる検索
* タグによる絞り込み
* 保存したURLの状態確認

## 使い方

保存したいWebページの情報をLinkVaultに登録します。

例：

```text
URL:
https://example.com/go-concurrency

タイトル:
Go Concurrency Patterns

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

状態: アクセス可能

Worker Poolについて学ぶ際の参考資料
```

## インストール方法

現在開発中のため、インストール方法は未定です。

開発の進行に合わせて追記します。

## 使用技術

### フロントエンド

* TypeScript

### バックエンド

* Go

### データベース

* 未定

その他の技術については、設計・開発の進行に合わせて決定します。

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
