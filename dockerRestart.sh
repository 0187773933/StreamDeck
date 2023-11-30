#!/bin/bash
APP_NAME="public-stream-deck-server"
id=$(sudo docker restart $APP_NAME)
sudo docker logs -f $id