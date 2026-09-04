#!/bin/sh
set -eu

base_dir=${RDB_BASE_DIR:-/home/root/.local/share/remarkable-customizations/daily-board}
config_file=${RDB_CONFIG_FILE:-$base_dir/device-sync.conf}
live_file=${RDB_LIVE_FILE:-/home/root/.local/share/remarkable-customizations/live/suspended.png}
state_dir=${RDB_STATE_DIR:-$base_dir/state}
lock_dir=$state_dir/update.lock
etag_file=$state_dir/etag
hash_file=$state_dir/sha256
download_file=$state_dir/suspended.download
headers_file=$state_dir/headers

log() {
    logger -t rich-daily-board "$*" 2>/dev/null || true
    printf '%s\n' "$*"
}

read_value() {
    key=$1
    sed -n "s/^${key}=//p" "$config_file" | sed -n '1p'
}

test -r "$config_file" || {
    log "configuration is missing: $config_file"
    exit 1
}
command -v wget >/dev/null 2>&1 || {
    log "wget is required"
    exit 1
}

board_url=$(read_value BOARD_URL)
board_token=$(read_value BOARD_TOKEN)
test -n "$board_url" || { log "BOARD_URL is empty"; exit 1; }
test -n "$board_token" || { log "BOARD_TOKEN is empty"; exit 1; }

mkdir -p "$state_dir" "$(dirname "$live_file")"
if ! mkdir "$lock_dir" 2>/dev/null; then
    log "an update is already running"
    exit 0
fi
trap 'rm -f "$download_file" "$headers_file"; rmdir "$lock_dir" 2>/dev/null || true' EXIT

etag=""
if [ -r "$etag_file" ]; then
    etag=$(sed -n '1p' "$etag_file")
fi

if [ -n "$etag" ]; then
    wget -q -S -T 30 \
        --header "Authorization: Bearer $board_token" \
        --header "If-None-Match: $etag" \
        -o "$headers_file" \
        -O "$download_file" \
        "$board_url" || wget_result=$?
else
    wget -q -S -T 30 \
        --header "Authorization: Bearer $board_token" \
        -o "$headers_file" \
        -O "$download_file" \
        "$board_url" || wget_result=$?
fi

status=$(awk '/  HTTP\// { code=$2 } END { print code }' "$headers_file")
if [ -z "$status" ]; then
    log "board endpoint is temporarily unreachable"
    exit 0
fi

case "$status" in
    304)
        log "board is unchanged"
        exit 0
        ;;
    200)
        ;;
    401|403)
        log "board endpoint rejected the device token (HTTP $status)"
        exit 1
        ;;
    *)
        log "board endpoint returned HTTP $status"
        exit 1
        ;;
esac

magic=$(hexdump -v -e '1/1 "%02x"' -n 8 "$download_file")
dimensions=$(hexdump -v -e '1/1 "%02x"' -s 16 -n 8 "$download_file")
size=$(wc -c < "$download_file" | tr -d ' ')
test "$magic" = "89504e470d0a1a0a" || {
    log "download rejected: invalid PNG signature"
    exit 1
}
test "$dimensions" = "000003ba000006a0" || {
    log "download rejected: expected 954 x 1696 PNG"
    exit 1
}
test "$size" -ge 10000 || {
    log "download rejected: PNG is unexpectedly small"
    exit 1
}

new_hash=$(sha256sum "$download_file" | cut -d' ' -f1)
old_hash=""
if [ -r "$hash_file" ]; then
    old_hash=$(sed -n '1p' "$hash_file")
fi
if [ "$new_hash" = "$old_hash" ]; then
    log "board content is unchanged"
    exit 0
fi

chmod 0644 "$download_file"
if [ "${RDB_SKIP_CHOWN:-0}" != "1" ]; then
    chown root:root "$download_file"
fi
mv -f "$download_file" "$live_file"
printf '%s\n' "$new_hash" > "$hash_file"

new_etag=$(awk 'tolower($1) == "etag:" { sub(/\r$/, "", $2); print $2 }' "$headers_file" | tail -n 1)
if [ -n "$new_etag" ]; then
    printf '%s\n' "$new_etag" > "$etag_file"
fi

sync
log "installed board $new_hash; it will appear on the next sleep transition"
