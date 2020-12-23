#!/usr/bin/env bash

set -Eeuxo pipefail

go build -o simplemon ./app/

systemctl stop simplemon || true
systemctl disable simplemon || true
systemctl daemon-reload

mkdir -p /srv/simplemon

cp simplemon /srv/simplemon/
cp conf/simplemon-conf.yml /srv/simplemon/

cp conf/simplemon.service /usr/lib/systemd/system/simplemon.service
systemctl daemon-reload
systemctl enable simplemon
