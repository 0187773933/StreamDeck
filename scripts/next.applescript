if application "Music" is running then
    tell application "Music"
        if player state is playing then
            next track
            return
        end if
    end tell
end if

if application "Spotify" is running then
    tell application "Spotify"
        if player state is playing then
            next track
            return
        end if
    end tell
end if