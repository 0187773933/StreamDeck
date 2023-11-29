#!/bin/bash
sudo chown morphs:morphs /dev/streamdeck
amixer sset Master 100%
amixer sset Master unmute

cd /home/morphs
sudo rm -rf /home/morphs/StreamDeck
git clone "https://github.com/0187773933/StreamDeck.git"
sudo chown -R morphs:morphs /home/morphs/StreamDeck
cd /home/morphs/StreamDeck
#exec bash
/usr/local/go/bin/go mod tidy
GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o /home/morphs/StreamDeck/server
exec /home/morphs/StreamDeck/server "$@"
