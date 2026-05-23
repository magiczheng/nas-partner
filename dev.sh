#!/usr/bin/env bash
set -e

ROOT_DIR="$(cd "$(dirname "$0")" && pwd)"

CLEAR=false
while getopts "c" opt; do
  case $opt in
    c) CLEAR=true ;;
    *) echo "用法: $0 [-c]" >&2; exit 1 ;;
  esac
done

DB_FILE="$ROOT_DIR/backend/data/nas-partner.db"

if [ "$CLEAR" = true ]; then
  echo "🧹 清除数据库: $DB_FILE"
  rm -f "$DB_FILE"
fi

cleanup() {
  echo ""
  echo "🛑 正在关闭服务..."
  kill $BACKEND_PID $FRONTEND_PID 2>/dev/null
  wait $BACKEND_PID $FRONTEND_PID 2>/dev/null
  echo "✅ 已退出"
}
trap cleanup EXIT INT TERM

echo "🚀 启动后端..."
cd "$ROOT_DIR/backend" && go run cmd/server/main.go &
BACKEND_PID=$!

sleep 2

echo "🚀 启动前端..."
cd "$ROOT_DIR/frontend" && npx vite --host &
FRONTEND_PID=$!

echo ""
echo "后端: http://localhost:8080"
echo "前端: http://localhost:5173"
echo "按 Ctrl+C 退出"
echo ""

wait
