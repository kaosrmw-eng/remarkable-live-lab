#!/bin/sh
set -eu

base_dir=/home/root/.local/share/remarkable-customizations/daily-board
unit_dir=/lib/systemd/system

test "$(id -u)" -eq 0
systemctl disable --now rich-daily-board-update.timer 2>/dev/null || true

mount -o remount,rw /
trap 'mount -o remount,ro /' EXIT
rm -f "$unit_dir/rich-daily-board-update.timer"
rm -f "$unit_dir/rich-daily-board-update.service"
systemctl daemon-reload
mount -o remount,ro /
trap - EXIT

rm -rf "$base_dir"
findmnt -n -o OPTIONS /
