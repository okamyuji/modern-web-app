---
title: "【Python】Slackの特定ユーザーメッセージを効率的に取得・分析するツール"
emoji: "💬"
type: "tech"
topics: ["python", "slack", "api", "automation", "データ分析"]
published: false
---

## はじめに

Slackでの会話データを分析したい、特定のメンバーの発言を抽出したい、そんなニーズはありませんか？

本記事では、**Slack APIを使って特定ユーザーのメッセージを効率的に取得するPythonツール**を作成しました。日付範囲指定、大量データ対応、セキュアな環境変数管理など、実用的な機能を網羅しています。

:::message
**📁 完全なソースコード**
この記事で紹介するツールの完全なソースコードは、以下のGitHubリポジトリで公開しています。
**🔗 [GitHub: slack-message-extractor](https://github.com/okamyuji/slack-my-conversation)**

すぐに試したい方は、リポジトリをクローンして使用してください。
:::

## 🎯 この記事で学べること

- Slack APIの2つの取得方法の使い分け（search.messages vs conversations.history）
- 環境変数を使ったセキュアな設定管理の実装方法
- 日付範囲指定とページネーション処理の実装パターン
- エラーハンドリングとトラブルシューティングのベストプラクティス
- Slack API活用の実践的なテクニック

## 📊 完成イメージ

```bash
$ python app.py

✅ 環境変数の設定を確認しました。
🔧 設定情報:
  チャンネルID: C1234567890
  ユーザーID: U1234567890
  トークン: xoxp-1234567890... (一部のみ表示)

=== 取得方法を選択してください ===
1. 直接検索（推奨）: search.messages APIで特定ユーザーのメッセージのみを取得
2. 全取得後フィルタ: 全メッセージを取得してからフィルタリング

選択してください (1/2): 1

=== 詳細設定（オプション）===
取得件数を指定してください（デフォルト: 100、最大値はAPI次第）: 50
開始日（省略可）: 2025-01-01
終了日（省略可）: 

検索結果: 23件のメッセージが見つかりました
```

## 🚀 機能概要

### 主要機能

| 機能 | 説明 | 対応状況 |
|------|------|----------|
| 特定ユーザー抽出 | 指定したユーザーのメッセージのみを取得 | ✅ |
| 日付範囲指定 | 2025-04-01以降など、期間を指定して取得 | ✅ |
| 2つの取得方法 | 効率的な直接検索 & 全取得後フィルタ | ✅ |
| 大量データ対応 | ページネーション機能で制限なく取得 | ✅ |
| JSONファイル保存 | 取得したメッセージをファイルに保存 | ✅ |
| 環境変数管理 | セキュアな設定管理（.env対応） | ✅ |
| 詳細エラーハンドリング | 分かりやすいエラーメッセージと解決方法 | ✅ |

### API制限と特徴比較

| API | 最大取得件数 | 特徴 | 推奨用途 |
|-----|-------------|------|----------|
| **search.messages** | 1,000件 | 効率的、直接検索 | 少量データ、特定条件 |
| **conversations.history** | 無制限 | ページネーション対応 | 大量データ、全履歴 |

## 🛠️ 実装手順

### 1. プロジェクト構成

```text
slack-my-conversation/
├── app.py              # メインアプリケーション
├── .env                # 環境変数（実際の値）
├── .env.sample         # 環境変数サンプル
├── requirements.txt    # 依存関係
├── .gitignore         # Git除外設定
└── README.md          # ドキュメント
```

### 2. 依存関係の設定

**requirements.txt**:

```txt
requests>=2.32.0
python-dotenv>=1.0.0
```

**インストール**:

```bash
pip install -r requirements.txt
```

### 3. 環境変数の設定

**.env.sample** （サンプルファイル）:

```env
# Slack API設定のサンプル
# このファイルをコピーして .env ファイルを作成し、実際の値を設定してください

# Slack APIトークン（Bot User OAuth Token）
# Slack App管理画面の「OAuth & Permissions」から取得
# 形式: xoxp-xxxxxxxxxx-xxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
SLACK_TOKEN=xoxp-your-slack-token-here

# SlackチャンネルID
# チャンネル名の右クリック → 「リンクをコピー」から取得可能
# 形式: C1234567890
SLACK_CHANNEL_ID=C1234567890

# SlackユーザーID（抽出対象のユーザー）
# プロフィールの「その他」→「メンバーIDをコピー」から取得可能
# 形式: U1234567890
SLACK_USER_ID=U1234567890
```

**.env** （実際の設定値の例）:

```env
SLACK_TOKEN=xoxp-1234567890-1234567890-1234567890-abcdef1234567890abcdef1234567890
SLACK_CHANNEL_ID=C1234567890
SLACK_USER_ID=U1234567890
```

### 4. 核心機能の実装解説

完全なソースコードは[GitHubリポジトリ](https://github.com/okamyuji/slack-my-conversation)で確認できますが、ここでは重要な実装パターンを抽出して解説します。

#### 🔐 セキュアな環境変数管理

```python
def validate_environment_variables() -> Dict[str, str]:
    """環境変数の検証とユーザーフレンドリーなエラーハンドリング"""
    load_dotenv()  # .envファイルを自動読み込み
    
    required_vars = {
        'SLACK_TOKEN': 'Slack APIトークン',
        'SLACK_CHANNEL_ID': 'Slackチャンネル ID',
        'SLACK_USER_ID': 'Slack ユーザー ID'
    }
    
    # 不足している環境変数を収集
    missing_vars = []
    config = {}
    
    for var_name, description in required_vars.items():
        value = os.getenv(var_name)
        if not value:
            missing_vars.append(f"  {var_name}: {description}")
        else:
            config[var_name] = value
    
    # エラー時は具体的な解決方法を提示
    if missing_vars:
        print("❌ エラー: 必要な環境変数が設定されていません。")
        print("\\n設定が必要な環境変数:")
        for var in missing_vars:
            print(var)
        print("\\n解決方法:")
        print("1. .envファイルを作成し、必要な値を設定してください")
        sys.exit(1)
    
    return config
```

**💡 ポイント:**

- `python-dotenv`で`.env`ファイルを自動読み込み
- 不足している環境変数を一括で確認
- エラー時に具体的な解決方法を提示

#### 📅 日付処理のユーティリティ

```python
def date_to_timestamp(date_str: str) -> Optional[float]:
    """人間が読みやすい日付をSlack APIが理解できる形式に変換"""
    try:
        # "2025-04-01" → "2025-04-01 00:00:00" に補完
        if len(date_str) == 10:
            date_str += " 00:00:00"
        
        dt = datetime.strptime(date_str, "%Y-%m-%d %H:%M:%S")
        dt = dt.replace(tzinfo=timezone.utc)
        return dt.timestamp()
    except ValueError as e:
        print(f"日付形式エラー: {e}")
        print("正しい形式: YYYY-MM-DD または YYYY-MM-DD HH:MM:SS")
        return None
```

**💡 ポイント:**

- 時刻省略時の自動補完機能
- エラー時の分かりやすいガイダンス
- タイムゾーン考慮済み

#### 🎯 方法1: 効率的な直接検索

```python
def search_user_messages(channel_id: str, user_id: str, token: str, 
                        count: int = 100, after: str = None, before: str = None):
    """search.messages APIで特定ユーザーのメッセージを直接取得"""
    
    # 検索クエリの動的構築
    query_parts = [f"in:<#{channel_id}>", f"from:<@{user_id}>"]
    
    if after:
        query_parts.append(f"after:{after}")
    if before:
        query_parts.append(f"before:{before}")
    
    query = " ".join(query_parts)
    
    params = {
        "query": query,
        "count": min(count, 1000),  # API制限を考慮
        "sort": "timestamp",
        "sort_dir": "desc"
    }
    
    # ... API呼び出し処理 ...
    
    if data['ok']:
        messages = data.get('messages', {}).get('matches', [])
        print(f"検索クエリ: {query}")
        print(f"検索結果: {len(messages)}件")
        return messages
```

**💡 ポイント:**

- 動的な検索クエリ構築で柔軟な条件指定
- API制限（1000件）を考慮した安全な実装
- 検索クエリの可視化でデバッグしやすい

#### 🔄 方法2: ページネーション対応の全取得

```python
def get_conversation_history(channel_id: str, token: str, limit: int = 100,
                           oldest: str = None, latest: str = None, get_all: bool = False):
    """conversations.history APIでページネーション対応の大量データ取得"""
    
    all_messages = []
    params = {"channel": channel_id, "limit": min(limit, 200)}
    
    # 日付範囲をタイムスタンプに変換
    if oldest:
        params["oldest"] = date_to_timestamp(oldest)
    if latest:
        params["latest"] = date_to_timestamp(latest)
    
    while True:
        response = requests.get(url, headers=headers, params=params)
        data = response.json()
        
        if data['ok']:
            messages = data['messages']
            all_messages.extend(messages)
            
            # ページネーション処理
            if (get_all and data.get('has_more') and 
                data.get('response_metadata', {}).get('next_cursor')):
                params['cursor'] = data['response_metadata']['next_cursor']
                print(f"取得中... 現在 {len(all_messages)} 件")
            else:
                break
        else:
            # エラーハンドリング...
            break
    
    return all_messages
```

**💡 ポイント:**

- `cursor`を使った正しいページネーション実装
- 進捗表示で大量データ取得時のUX向上
- API制限（200件/回）を考慮した設計

## 🔧 Slack App設定手順

### 1. Slack Appの作成

1. [Slack API Dashboard](https://api.slack.com/apps)にアクセス
2. 「Create New App」をクリック
3. 「From scratch」を選択
4. アプリ名とワークスペースを設定

### 2. 必要なスコープの追加

#### 方法1（推奨）を使用する場合

- `search:read`

#### 方法2を使用する場合

- `channels:history`
- `groups:history`
- `im:history`
- `mpim:history`

**設定手順:**

1. 左メニューから「OAuth & Permissions」を選択
2. 「Scopes」セクションで「User Token Scopes」に上記スコープを追加
3. 「Install to Workspace」をクリック
4. 権限を承認

### 3. 設定値の取得

#### APIトークン

「OAuth & Permissions」画面の「User OAuth Token」をコピー

#### チャンネルID

ブラウザ版Slackでチャンネルを開き、URLから取得

```text
https://app.slack.com/client/T.../C1234567890
                                ↑このID部分
```

#### ユーザーID

Slackでユーザープロフィールを開き、「その他」→「メンバーIDをコピー」

## 💡 使用方法とオプション

### 基本的な実行

```bash
python app.py
```

### 実行時の選択肢

#### 1. 取得方法の選択

```text
=== 取得方法を選択してください ===
1. 直接検索（推奨）: search.messages APIで特定ユーザーのメッセージのみを取得
2. 全取得後フィルタ: 全メッセージを取得してからフィルタリング

選択してください (1/2): 1
```

#### 2. 詳細設定

```text
=== 詳細設定（オプション）===
取得件数を指定してください（デフォルト: 100、最大値はAPI次第）: 500

日付範囲を指定できます（例: 2025-04-01）:
開始日（省略可）: 2025-01-01
終了日（省略可）: 2025-03-31
```

#### 3. 保存オプション

```text
メッセージをJSONファイルに保存しますか？ (y/n): y
メッセージを U1234567890_slack_messages_C1234567890.json に保存しました。
```

#### 出力例

```text
ユーザー U1234567890 のメッセージ (23件):

1. User: U1234567890
   Time: 2025-01-15 14:30:25 (1705317025.123456)
   Text: プロジェクトの進捗はいかがですか？
----------------------------------------------------------------------

2. User: U1234567890
   Time: 2025-01-15 10:15:30 (1705301730.654321)
   Text: おはようございます！今日もよろしくお願いします。
----------------------------------------------------------------------
```

## 🐛 トラブルシューティング

### よくあるエラーと解決方法

#### 1. 「missing_scope」エラー

```text
❌ Slack API Error: missing_scope

=== 解決方法 ===
Slack APIトークンに以下のスコープが必要です:
- search:read (メッセージ検索用)

Slack App管理画面でこのスコープを追加し、
ワークスペースに再インストールしてください。
```

**解決手順:**

1. Slack App管理画面の「OAuth & Permissions」を開く
2. 必要なスコープを追加
3. 「Reinstall to Workspace」をクリック

#### 2. 「channel_not_found」エラー

```text
❌ Slack API Error: channel_not_found
チャンネルID 'C1234567890' が見つかりません。
```

**解決手順:**

1. チャンネルIDが正しいか確認
2. Botがチャンネルのメンバーか確認
3. チャンネルが存在するか確認

#### 3. 環境変数エラー

```text
❌ エラー: 必要な環境変数が設定されていません。

設定が必要な環境変数:
  SLACK_TOKEN: Slack APIトークン
  SLACK_CHANNEL_ID: Slackチャンネル ID
  SLACK_USER_ID: Slack ユーザー ID
```

**解決手順:**

1. `.env`ファイルが存在するか確認
2. 必要な環境変数が全て設定されているか確認
3. `.env.sample`を参考に正しい形式で設定

## 📈 応用例とカスタマイズ

ここでは、基本ツールを拡張するための実用的なパターンを紹介します。

### 1. 複数ユーザーの一括取得

```python
def get_multiple_users_messages(channel_id: str, user_ids: List[str], token: str):
    """複数ユーザーのメッセージを一括取得"""
    all_results = {}
    
    for user_id in user_ids:
        print(f"ユーザー {user_id} のメッセージを取得中...")
        messages = search_user_messages(channel_id, user_id, token)
        all_results[user_id] = messages
        time.sleep(1)  # レート制限対策
        
    return all_results

# 使用例
user_list = ['U1234567890', 'U12345678', 'U87654321']
results = get_multiple_users_messages(channel_id, user_list, token)
```

### 2. 感情分析との連携

```python
from textblob import TextBlob

def analyze_message_sentiment(messages: List[Dict]):
    """メッセージの感情分析とキーワード抽出"""
    analysis_results = {
        'positive': 0,
        'negative': 0,
        'neutral': 0,
        'keywords': {},
        'total': len(messages)
    }
    
    for message in messages:
        text = message.get('text', '')
        if not text:
            continue
            
        # 感情分析
        blob = TextBlob(text)
        sentiment = blob.sentiment.polarity
        
        if sentiment > 0.1:
            analysis_results['positive'] += 1
        elif sentiment < -0.1:
            analysis_results['negative'] += 1
        else:
            analysis_results['neutral'] += 1
        
        # キーワード抽出（名詞のみ）
        for word, pos in blob.tags:
            if pos.startswith('NN'):  # 名詞
                analysis_results['keywords'][word] = analysis_results['keywords'].get(word, 0) + 1
    
    return analysis_results
```

### 3. 定期実行とデータ蓄積パターン

```python
import schedule
import time
from datetime import datetime, timedelta

def daily_message_collection():
    """日次でメッセージを収集してCSVに保存"""
    today = datetime.now().strftime('%Y-%m-%d')
    yesterday = (datetime.now() - timedelta(days=1)).strftime('%Y-%m-%d')
    
    messages = search_user_messages(
        channel_id, user_id, token, 
        after=yesterday, before=today
    )
    
    if messages:
        # CSVファイルに追記
        import csv
        filename = f"daily_messages_{user_id}.csv"
        
        with open(filename, 'a', newline='', encoding='utf-8') as f:
            writer = csv.writer(f)
            for msg in messages:
                writer.writerow([
                    msg.get('ts', ''),
                    msg.get('user', ''),
                    msg.get('text', '').replace('\n', ' ')
                ])
        
        print(f"{yesterday}のメッセージ {len(messages)}件を保存しました")

# スケジュール設定
schedule.every().day.at("09:00").do(daily_message_collection)
```

## 🔒 セキュリティ考慮事項

### 1. 環境変数の管理

```bash
# .gitignoreに必ず追加
echo ".env" >> .gitignore
```

### 2. トークンの権限最小化

- 必要最小限のスコープのみを付与
- 定期的なトークンのローテーション
- 使用しないトークンの無効化

### 3. ログ出力の注意

```python
# ❌ 危険: トークンがログに出力される
print(f"Token: {token}")

# ✅ 安全: 一部のみ表示
print(f"Token: {token[:20]}...")
```

## 📊 パフォーマンス最適化

### API呼び出し頻度の制御

```python
import time
from functools import wraps

def rate_limit(calls_per_minute=60):
    """APIレート制限デコレータ"""
    min_interval = 60.0 / calls_per_minute
    last_called = [0.0]
    
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            elapsed = time.time() - last_called[0]
            left_to_wait = min_interval - elapsed
            if left_to_wait > 0:
                time.sleep(left_to_wait)
            ret = func(*args, **kwargs)
            last_called[0] = time.time()
            return ret
        return wrapper
    return decorator

@rate_limit(calls_per_minute=50)  # 1分間に50回まで
def api_call_with_rate_limit():
    # API呼び出し処理
    pass
```

## 🎉 まとめ

本記事では、Slack APIを使った特定ユーザーメッセージ取得ツールを作成しました。

### 🌟 主要なポイント

1. **2つのAPI手法**: 効率的な直接検索と大量データ対応の全取得
2. **セキュアな設定管理**: 環境変数とバリデーション機能
3. **実用的な機能**: 日付範囲指定、ページネーション、エラーハンドリング
4. **拡張性**: カスタマイズや他システムとの連携が容易

### 🚀 今後の発展可能性

- **データ分析**: 感情分析、キーワード抽出、トレンド分析
- **自動化**: 定期実行、アラート機能、レポート生成
- **可視化**: グラフ作成、ダッシュボード構築
- **AI連携**: ChatGPT APIとの組み合わせで要約・分析

このツールを基盤として、様々なSlackデータ活用にチャレンジしてみてください！

### 📚 参考リンク

- [Slack API Documentation](https://api.slack.com/)
- [Python requests library](https://docs.python-requests.org/)
- [python-dotenv](https://pypi.org/project/python-dotenv/)

---

もし記事が役に立ったら、ぜひ**いいね**や**コメント**をお願いします！質問や改善提案もお待ちしております 🙌
