#!/bin/bash
sudo chown morphs:morphs /dev/streamdeck
amixer sset Master 100%
amixer sset Master unmute
exec /home/morphs/StreamDeckServer/server "$@"
#exec bash
