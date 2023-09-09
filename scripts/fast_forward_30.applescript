-- property skip_secs : 30 -- use negative value to rewind, positive value to skip ahead
-- https://jaguchi.com/blog/2012/03/itunes-30-second-skip-podcasts-audiobooks-only/
-- https://dougscripts.com/itunes/itinfo/203changes.php
-- https://github.com/dronir/SpotifyControl
-- https://github.com/andrehaveman/spotify-node-applescript/tree/master/lib/scripts
-- https://github.com/shaundon/hubot-osx-control/blob/master/tv-server/SpotifyControl.scpt
property skip_secs : 30
if application "Music" is running then
	tell application "Music"
		if (exists of current track) is true then
			set target_time to (player position + skip_secs)
			if (target_time > finish of current track) then
				-- next track
			else if (target_time < 3) then
				-- back track
			else
				set player position to target_time
			end if
		end if
	end tell
end if
if application "Spotify" is running then
	tell application "Spotify"
		set spotify_state to (player state as text)
		if spotify_state is "playing" then
			log spotify_state
			set current_tracks_position to player position
			set current_tracks_duration to duration of current track
			set target_time to (current_tracks_position + skip_secs)
			log current_tracks_position
			log current_tracks_duration
			log target_time
			if (target_time > current_tracks_duration) then
				-- next track
				log "target time was greater"
			else if (target_time < 3) then
				-- back track
			else
				set player position to target_time
			end if
		end if
	end tell
end if