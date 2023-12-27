#!/usr/bin/env python3

from StreamDeck.DeviceManager import DeviceManager
# python3 -m pip install streamdeck
# https://github.com/abcminiuser/python-elgato-streamdeck
# https://github.com/abcminiuser/python-elgato-streamdeck/blob/master/test/test.py
# https://python-elgato-streamdeck.readthedocs.io/en/stable/modules/devices.html#StreamDeck.Devices.StreamDeck.StreamDeck

if __name__ == "__main__":
	x = DeviceManager( transport=None )
	devices = x.enumerate()
	print( devices )
	for i , d in enumerate( devices ):
		with d:
			d.open()
			print( i , "===" , d.get_serial_number() )
			d.close()