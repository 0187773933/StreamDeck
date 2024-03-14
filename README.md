# Stream Deck Controller

## find sound card and update asound.conf

`aplay -l`

## osx 14.0 broke hid stuff in meusli/streamdeck somehow

- you can still get it with https://github.com/dh1tw/hid
- but now we have to figure out a way to patch it into meusli/streamdeck
- https://github.com/muesli/streamdeck/blob/v0.4.0/streamdeck.go
- https://github.com/karalabe/hid

- https://github.com/karalabe/hid/issues/35
- https://github.com/karalabe/hid/pull/33/commits/4cde0f992c7cd347cf9d819f9ff6178831dbebac
- https://raw.githubusercontent.com/karalabe/hid/e6a971e4cb40e37140f0de29e039d8832d37614b/upgrade-hidapi.sh

## ubuntu

- sudo apt-get install libasound2-dev
- sudo apt-get install libhidapi-dev
- sudo usermod -a -G dialout morphs
- sudo usermod -a -G audio morphs
- sudo nano /etc/udev/rules.d/99-streamdeck.rules
	- SUBSYSTEM=="usb", ATTRS{idVendor}=="0fd9", ATTRS{idProduct}=="006d", MODE="0666", GROUP="morphs", SYMLINK+="streamdeck"
- sudo udevadm control --reload-rules && sudo udevadm trigger

## misc

- https://www.amazon.com/gp/product/B072BMG9TB

## Todo

- Twilio Call Support
- routes
	- clear
	- turn off
	- mute / toggle audio responses
	- get logs