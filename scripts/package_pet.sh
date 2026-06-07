#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

DIST="${DIST:-dist}"
PET_NAME="${PET_NAME:-lucheng-pet}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
LDFLAGS="-s -w"

usage() {
	cat <<EOF
用法: $0 <darwin|windows|all>

  darwin   打包 macOS 可执行文件与 .app
  windows  打包 Windows .exe（可在 macOS/Linux 上交叉编译）
  all      当前系统为 macOS 时同时打 darwin + windows
EOF
}

mkdir -p "$DIST"

build_darwin() {
	local arch="${1:-$(uname -m)}"
	case "$arch" in
		arm64 | aarch64) arch=arm64 ;;
		x86_64 | amd64) arch=amd64 ;;
		*) echo "不支持的 macOS 架构: $arch" >&2; exit 1 ;;
	esac

	local bin="$DIST/${PET_NAME}-darwin-${arch}"
	echo "==> 构建 macOS ${arch}"
	CGO_ENABLED=1 GOOS=darwin GOARCH="$arch" go build -tags ebitenpet -ldflags "$LDFLAGS" -o "$bin" .

	local app="$DIST/${PET_NAME}.app"
	rm -rf "$app"
	mkdir -p "$app/Contents/MacOS"
	cp "$bin" "$app/Contents/MacOS/lucheng-pet"
	chmod +x "$app/Contents/MacOS/lucheng-pet"
	sed "s/__VERSION__/${VERSION}/g" packaging/darwin/Info.plist > "$app/Contents/Info.plist"
	echo "==> 已生成 $bin"
	echo "==> 已生成 $app"
}

build_windows() {
	local bin="$DIST/${PET_NAME}-windows-amd64.exe"
	echo "==> 构建 Windows amd64"
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags ebitenpet -ldflags "$LDFLAGS" -o "$bin" .
	echo "==> 已生成 $bin"
}

cmd="${1:-}"
case "$cmd" in
	darwin) build_darwin ;;
	windows) build_windows ;;
	all)
		build_darwin
		build_windows
		;;
	-h | --help | help)
		usage
		;;
	*)
		usage >&2
		exit 1
		;;
esac
