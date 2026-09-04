#!/bin/sh
set -eu
base=/home/root/gomarkable-move-test
test -x "$base/gomarkablestream-share"
test -f "$base/gomarkable-share.service"
test -f "$base/gomarkable-private.env"
# /etc is volatile on this firmware: store unit and boot link in /lib instead.
test ! -e /lib/systemd/system/gomarkable-share.service
mount -o remount,rw /
trap 'mount -o remount,ro /' EXIT
cp "$base/gomarkable-share.service" /lib/systemd/system/gomarkable-share.service
chmod 0644 /lib/systemd/system/gomarkable-share.service
mkdir -p /lib/systemd/system/multi-user.target.wants
ln -s ../gomarkable-share.service /lib/systemd/system/multi-user.target.wants/gomarkable-share.service
mount -o remount,ro /
trap - EXIT
systemctl daemon-reload
systemctl stop gomarkable-public-test.service gomarkable-split-test.service gomarkable-move-controls.service || true
systemctl start gomarkable-share.service
systemctl show gomarkable-share.service -p ActiveState -p RuntimeMaxUSec -p UnitFileState
