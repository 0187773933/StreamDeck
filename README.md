# Stream Deck Controller

## fucking osx 14.0 broke hid stuff in meusli/streamdeck somehow

- you can still get it with https://github.com/dh1tw/hid
- but now we have to figure out a way to patch it into meusli/streamdeck
- https://github.com/muesli/streamdeck/blob/v0.4.0/streamdeck.go
- https://github.com/karalabe/hid

## Todo

- Twilio Call Support
- routes
	- clear
	- turn off
	- mute / toggle audio responses
	- get logs
- global cooldown stuff
- Redudency Checks
	- if stream deck USB became un-plugged / re-plugged