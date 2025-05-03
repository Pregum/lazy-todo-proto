# Lazy Todo

A simple and efficient terminal-based todo application with a modern UI.

シンプルで効率的なターミナルベースのタスク管理アプリケーション。

## Features / 機能

- 📝 Simple task management / シンプルなタスク管理
- 🎨 Modern terminal UI / モダンなターミナルUI
- 🔍 Quick search / クイック検索
- 🔄 Undo/Redo support / 元に戻す/やり直し機能
- 📂 Task persistence / タスクの永続化

## Project Structure / プロジェクト構成

```shell
.
├── cmd/
│   └── lazytodo/          # メインエントリーポイント
│       └── main.go
├── internal/
│   ├── app/               # アプリケーションロジック
│   │   └── model.go
│   ├── domain/            # ドメインモデル
│   │   └── task.go
│   └── infrastructure/    # 永続化層
│       └── storage.go
├── pkg/
│   └── ui/               # UI関連のユーティリティ
│       └── styles.go
├── go.mod
├── go.sum
└── README.md
```

## Installation / インストール

### Quick Install / 簡単なインストール方法

```bash
# 1コマンドでインストール
go install github.com/Pregum/lazy-todo-proto/cmd/lazytodo@latest
```

### Manual Install / 手動インストール

```bash
# Clone the repository
git clone https://github.com/Pregum/lazy-todo-proto.git
cd lazy-todo-proto

# Build the application
go build -o lazytodo ./cmd/lazytodo

# Install the application
mkdir -p $GOPATH/bin
cp lazytodo $GOPATH/bin/

# Add GOPATH/bin to your PATH (if not already added)
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.zshrc  # for zsh
# or
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc  # for bash

# Reload your shell configuration
source ~/.zshrc  # for zsh
# or
source ~/.bashrc  # for bash
```

### Usage / 使用方法

```bash
# Run the application
./lazytodo  # macOSの場合
# or
lazytodo    # PATHが設定されている場合
```

## Usage / 使い方

### Basic Commands / 基本コマンド

| Key | Action | 操作 |
|-----|--------|------|
| `n` | New task | 新規タスク |
| `e` | Edit task | タスクの編集 |
| `Space` | Toggle task status | タスクの状態を切り替え |
| `d` | Delete task | タスクの削除 |
| `↑/k` | Move up | 上に移動 |
| `↓/j` | Move down | 下に移動 |
| `Tab/h/l` | Switch pane | ペインの切り替え |
| `q` | Quit | 終了 |

### Filtering / フィルタリング

| Key | Action | 操作 |
|-----|--------|------|
| `a` | Show all tasks | すべてのタスクを表示 |
| `t` | Show active tasks | アクティブなタスクを表示 |
| `c` | Show completed tasks | 完了したタスクを表示 |
| `/` | Search mode | 検索モード |
| `Esc` | Clear search | 検索をクリア |

### History / 履歴

| Key | Action | 操作 |
|-----|--------|------|
| `u` | Undo | 元に戻す |
| `r` | Redo | やり直し |

## Search Mode / 検索モード

In search mode, you can use the following commands:
検索モードでは、以下のコマンドが使用できます：

- `Backspace`/`Delete`: Delete one character / 1文字削除
- `Ctrl+w`: Delete one word / 1単語削除
- `Ctrl+u`: Clear search text / 検索テキストをクリア
- `Enter`: Exit search mode / 検索モードを終了
- `Esc`: Clear search and exit / 検索をクリアして終了

## License / ライセンス

MIT License
