#!/bin/bash

# on this host machine , where you run dockerRuns.sh
# sudo nano /etc/udev/rules.d/99-streamdeck.rules
	#  SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="006d", MODE="0666", GROUP="morphs", SYMLINK+="streamdeck"
# sudo udevadm control --reload-rules && sudo udevadm trigger
# ls -l /dev/streamdeck
	# lrwxrwxrwx 1 root root 15 Nov 28 11:19 /dev/streamdeck -> bus/usb/001/006

APP_NAME="public-stream-deck-server"
sudo docker rm -f $APP_NAME || echo ""
#sudo docker run --device=/dev/streamdeck -it $APP_NAME
id=$(sudo docker run -dit \
--name $APP_NAME \
--restart='always' \
--privileged \
--device=/dev/snd \
-v $(pwd)/SAVE_FILES:/home/morphs/SAVE_FILES:rw \
--mount type=bind,source="$(pwd)"/config.yaml,target=/home/morphs/StreamDeckServer/config.yaml \
-p 5953:5953 \
$APP_NAME config.yaml)
sudo docker logs -f $id


#--device=/dev/streamdeck \
