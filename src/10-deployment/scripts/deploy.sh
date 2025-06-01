#!/bin/bash

# デプロイメントスクリプト
# 使用方法: ./scripts/deploy.sh [version] [environment]

set -e

# 色付きログ出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 変数設定
APP_NAME="deployment-demo"
VERSION=${1:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev-$(date +%Y%m%d-%H%M%S)")}
ENVIRONMENT=${2:-"production"}
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# デプロイ設定
DEPLOY_USER="${DEPLOY_USER:-deploy}"
DEPLOY_HOST="${DEPLOY_HOST:-localhost}"
DEPLOY_PATH="${DEPLOY_PATH:-/var/www/${APP_NAME}}"
BACKUP_PATH="${DEPLOY_PATH}/backups"
HEALTH_CHECK_URL="${HEALTH_CHECK_URL:-http://localhost:8080/health}"
HEALTH_CHECK_TIMEOUT=${HEALTH_CHECK_TIMEOUT:-30}
ROLLBACK_TIMEOUT=${ROLLBACK_TIMEOUT:-60}

# 一時ポート（Blue-Greenデプロイ用）
TEMP_PORT=${TEMP_PORT:-8081}

log_info "=== デプロイメント開始 ==="
log_info "アプリケーション: ${APP_NAME}"
log_info "バージョン: ${VERSION}"
log_info "環境: ${ENVIRONMENT}"
log_info "デプロイ先: ${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}"

# 前提条件チェック
check_prerequisites() {
    log_info "前提条件をチェック中..."

    # Gitの状態確認
    if [ "$ENVIRONMENT" = "production" ] && [ -d ".git" ]; then
        if ! git diff-index --quiet HEAD --; then
            log_warning "未コミットの変更があります"
            read -p "続行しますか? (y/N): " -r
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                log_error "デプロイを中止しました"
                exit 1
            fi
        fi
    fi

    # 必要なコマンドの確認
    for cmd in git make curl ssh; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "$cmd コマンドが見つかりません"
            exit 1
        fi
    done

    # SSH接続確認
    if ! ssh -o ConnectTimeout=5 -o BatchMode=yes "${DEPLOY_USER}@${DEPLOY_HOST}" exit 2>/dev/null; then
        log_error "SSH接続に失敗しました: ${DEPLOY_USER}@${DEPLOY_HOST}"
        exit 1
    fi

    log_success "前提条件チェック完了"
}

# ビルド実行
build_application() {
    log_info "アプリケーションをビルド中..."

    # クリーンビルド
    make clean

    # テスト実行
    if [ "$ENVIRONMENT" = "production" ]; then
        log_info "テストを実行中..."
        make test || {
            log_error "テストが失敗しました"
            exit 1
        }
    fi

    # ビルド実行
    VERSION="$VERSION" BUILD_TIME="$BUILD_TIME" make build || {
        log_error "ビルドが失敗しました"
        exit 1
    }

    # バイナリの確認
    if [ ! -f "bin/${APP_NAME}" ]; then
        log_error "ビルドされたバイナリが見つかりません"
        exit 1
    fi

    log_success "ビルド完了"
}

# ファイル転送
transfer_files() {
    log_info "ファイルを転送中..."

    # 転送先ディレクトリの準備
    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "
        mkdir -p ${DEPLOY_PATH}/{bin,logs,backups}
        mkdir -p ${DEPLOY_PATH}/static/{css,js,images}
        mkdir -p ${DEPLOY_PATH}/internal/templates
    "

    # バイナリファイルの転送
    scp "bin/${APP_NAME}" "${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/bin/${APP_NAME}-${VERSION}"

    # 静的ファイルの転送
    if [ -d "static" ]; then
        rsync -avz --delete static/ "${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/static/"
    fi

    # テンプレートファイルの転送
    if [ -d "internal/templates" ]; then
        rsync -avz --delete internal/templates/ "${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/internal/templates/"
    fi

    # 設定ファイルの転送
    if [ -f ".env.${ENVIRONMENT}" ]; then
        scp ".env.${ENVIRONMENT}" "${DEPLOY_USER}@${DEPLOY_HOST}:${DEPLOY_PATH}/.env"
    fi

    # 実行権限の設定
    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "chmod +x ${DEPLOY_PATH}/bin/${APP_NAME}-${VERSION}"

    log_success "ファイル転送完了"
}

# ヘルスチェック実行
health_check() {
    local url=$1
    local timeout=$2
    local start_time=$(date +%s)

    log_info "ヘルスチェック実行中: $url"

    while [ $(($(date +%s) - start_time)) -lt $timeout ]; do
        if curl -f -s "$url" > /dev/null; then
            log_success "ヘルスチェック成功"
            return 0
        fi
        sleep 2
    done

    log_error "ヘルスチェックタイムアウト"
    return 1
}

# Blue-Greenデプロイメント
blue_green_deploy() {
    log_info "Blue-Greenデプロイメントを実行中..."

    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << EOF
        set -e
        cd ${DEPLOY_PATH}

        # 現在の実行中プロセスの確認
        CURRENT_PID=\$(pgrep -f "${APP_NAME}" || echo "")
        if [ -n "\$CURRENT_PID" ]; then
            echo "現在のプロセス: \$CURRENT_PID"
        fi

        # バックアップの作成
        if [ -f "bin/${APP_NAME}" ]; then
            BACKUP_FILE="backups/${APP_NAME}-\$(date +%Y%m%d-%H%M%S)"
            cp "bin/${APP_NAME}" "\$BACKUP_FILE"
            echo "バックアップ作成: \$BACKUP_FILE"
        fi

        # 新しいバイナリの配置
        cp "bin/${APP_NAME}-${VERSION}" "bin/${APP_NAME}-new"

        # 新しいインスタンスを一時ポートで起動
        echo "新しいインスタンスを起動中... (ポート: ${TEMP_PORT})"
        ENV=${ENVIRONMENT} PORT=${TEMP_PORT} ./bin/${APP_NAME}-new > logs/app-new.log 2>&1 &
        NEW_PID=\$!
        echo "新しいプロセス: \$NEW_PID"

        # 起動待機
        sleep 5

        # 新しいインスタンスのプロセス確認
        if ! kill -0 \$NEW_PID 2>/dev/null; then
            echo "新しいインスタンスの起動に失敗しました"
            exit 1
        fi
EOF

    # 新しいインスタンスのヘルスチェック
    local temp_health_url="http://${DEPLOY_HOST}:${TEMP_PORT}/health"
    if ! health_check "$temp_health_url" $HEALTH_CHECK_TIMEOUT; then
        log_error "新しいインスタンスのヘルスチェックに失敗"
        ssh "${DEPLOY_USER}@${DEPLOY_HOST}" "pkill -f '${APP_NAME}-new' || true"
        return 1
    fi

    # 切り替え実行
    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << EOF
        set -e
        cd ${DEPLOY_PATH}

        # 古いプロセスの停止
        if [ -n "\$CURRENT_PID" ]; then
            echo "古いプロセスを停止中: \$CURRENT_PID"
            kill -TERM \$CURRENT_PID 2>/dev/null || true
            
            # グレースフル停止の待機
            for i in {1..30}; do
                if ! kill -0 \$CURRENT_PID 2>/dev/null; then
                    break
                fi
                sleep 1
            done
            
            # 強制停止
            kill -KILL \$CURRENT_PID 2>/dev/null || true
        fi

        # 新しいプロセスを停止して本番ポートで再起動
        NEW_PID=\$(pgrep -f "${APP_NAME}-new" || echo "")
        if [ -n "\$NEW_PID" ]; then
            kill -TERM \$NEW_PID 2>/dev/null || true
            sleep 2
        fi

        # 本番用バイナリに置き換え
        mv "bin/${APP_NAME}-new" "bin/${APP_NAME}"

        # 本番ポートで起動
        echo "本番インスタンスを起動中... (ポート: 8080)"
        ENV=${ENVIRONMENT} PORT=8080 ./bin/${APP_NAME} > logs/app.log 2>&1 &
        MAIN_PID=\$!
        echo "本番プロセス: \$MAIN_PID"

        # PIDファイルに保存
        echo \$MAIN_PID > ${APP_NAME}.pid
EOF

    # 本番インスタンスのヘルスチェック
    if ! health_check "$HEALTH_CHECK_URL" $HEALTH_CHECK_TIMEOUT; then
        log_error "本番インスタンスのヘルスチェックに失敗"
        rollback
        return 1
    fi

    log_success "Blue-Greenデプロイメント完了"
}

# ローリングアップデート（複数インスタンス用）
rolling_update() {
    log_info "ローリングアップデートを実行中..."
    
    # 実装例（複数サーバーの場合）
    local servers=("${DEPLOY_HOST}")
    
    for server in "${servers[@]}"; do
        log_info "サーバー $server を更新中..."
        
        # ロードバランサーから除外
        # curl -X POST "http://loadbalancer/api/servers/$server/disable"
        
        # デプロイ実行
        DEPLOY_HOST="$server" blue_green_deploy
        
        # ロードバランサーに復帰
        # curl -X POST "http://loadbalancer/api/servers/$server/enable"
        
        log_success "サーバー $server の更新完了"
    done
}

# ロールバック
rollback() {
    log_warning "ロールバックを実行中..."

    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << EOF
        set -e
        cd ${DEPLOY_PATH}

        # 現在のプロセスを停止
        if [ -f "${APP_NAME}.pid" ]; then
            PID=\$(cat ${APP_NAME}.pid)
            if kill -0 \$PID 2>/dev/null; then
                kill -TERM \$PID
                sleep 5
                kill -KILL \$PID 2>/dev/null || true
            fi
        else
            pkill -f "${APP_NAME}" || true
        fi

        # 最新のバックアップを復元
        LATEST_BACKUP=\$(ls -t backups/${APP_NAME}-* 2>/dev/null | head -1 || echo "")
        if [ -n "\$LATEST_BACKUP" ]; then
            echo "バックアップから復元: \$LATEST_BACKUP"
            cp "\$LATEST_BACKUP" "bin/${APP_NAME}"
            
            # 復元したバイナリで起動
            ENV=${ENVIRONMENT} PORT=8080 ./bin/${APP_NAME} > logs/app.log 2>&1 &
            echo \$! > ${APP_NAME}.pid
            
            echo "ロールバック完了"
        else
            echo "バックアップファイルが見つかりません"
            exit 1
        fi
EOF

    if health_check "$HEALTH_CHECK_URL" $ROLLBACK_TIMEOUT; then
        log_success "ロールバック成功"
    else
        log_error "ロールバック失敗"
        exit 1
    fi
}

# デプロイ後の確認
post_deploy_check() {
    log_info "デプロイ後チェックを実行中..."

    # バージョン確認
    local deployed_version
    deployed_version=$(curl -s "${HEALTH_CHECK_URL}" | grep -o '"version":"[^"]*"' | cut -d'"' -f4 || echo "unknown")
    
    if [ "$deployed_version" = "$VERSION" ]; then
        log_success "バージョン確認OK: $deployed_version"
    else
        log_warning "バージョンが一致しません。期待値: $VERSION, 実際: $deployed_version"
    fi

    # 基本的な動作確認
    local endpoints=("/health" "/metrics")
    for endpoint in "${endpoints[@]}"; do
        local url="http://${DEPLOY_HOST}:8080${endpoint}"
        if curl -f -s "$url" > /dev/null; then
            log_success "エンドポイント確認OK: $endpoint"
        else
            log_warning "エンドポイント確認NG: $endpoint"
        fi
    done

    log_success "デプロイ後チェック完了"
}

# クリーンアップ
cleanup() {
    log_info "クリーンアップ中..."

    ssh "${DEPLOY_USER}@${DEPLOY_HOST}" << EOF
        cd ${DEPLOY_PATH}

        # 古いバックアップファイルの削除（7日以上前）
        find backups/ -name "${APP_NAME}-*" -mtime +7 -delete 2>/dev/null || true

        # 古いログファイルの削除（30日以上前）
        find logs/ -name "*.log" -mtime +30 -delete 2>/dev/null || true

        # 一時ファイルの削除
        rm -f bin/${APP_NAME}-new bin/${APP_NAME}-${VERSION} 2>/dev/null || true
EOF

    log_success "クリーンアップ完了"
}

# メイン実行
main() {
    check_prerequisites
    build_application
    transfer_files
    
    # デプロイ戦略の選択
    case "${DEPLOY_STRATEGY:-blue-green}" in
        "blue-green")
            blue_green_deploy
            ;;
        "rolling")
            rolling_update
            ;;
        *)
            log_error "不明なデプロイ戦略: ${DEPLOY_STRATEGY}"
            exit 1
            ;;
    esac
    
    post_deploy_check
    cleanup
    
    log_success "=== デプロイメント完了 ==="
    log_info "バージョン: $VERSION"
    log_info "環境: $ENVIRONMENT"
    log_info "URL: $HEALTH_CHECK_URL"
}

# エラーハンドリング
trap 'log_error "デプロイメントが失敗しました"; exit 1' ERR

# スクリプトが直接実行された場合のみmainを実行
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi