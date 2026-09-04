#!/bin/sh
set -eu

source_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
base_dir=/home/root/.local/share/remarkable-customizations/daily-board
unit_dir=/lib/systemd/system

test "$(id -u)" -eq 0
test -f "$source_dir/device-sync.conf"
grep -q '^BOARD_URL=https://' "$source_dir/device-sync.conf"
grep -q '^BOARD_TOKEN=.' "$source_dir/device-sync.conf"
command -v wget >/dev/null 2>&1
command -v hexdump >/dev/null 2>&1

mkdir -p "$base_dir/state"
cp "$source_dir/update-board.sh" "$base_dir/update-board.sh"
cp "$source_dir/device-sync.conf" "$base_dir/device-sync.conf"
chmod 0755 "$base_dir/update-board.sh"
chmod 0600 "$base_dir/device-sync.conf"
chown root:root "$base_dir/update-board.sh" "$base_dir/device-sync.conf"

mount -o remount,rw /
trap 'mount -o remount,ro /' EXIT
cp "$source_dir/rich-daily-board-update.service" "$unit_dir/rich-daily-board-update.service"
cp "$source_dir/rich-daily-board-update.timer" "$unit_dir/rich-daily-board-update.timer"
chmod 0644 "$unit_dir/rich-daily-board-update.service" "$unit_dir/rich-daily-board-update.timer"
chown root:root "$unit_dir/rich-daily-board-update.service" "$unit_dir/rich-daily-board-update.timer"
systemctl daemon-reload
systemctl enable rich-daily-board-update.timer
mount -o remount,ro /
trap - EXIT

systemctl start rich-daily-board-update.timer
systemctl start rich-daily-board-update.service
systemctl --no-pager status rich-daily-board-update.timer || true
findmnt -n -o OPTIONS /
