package types

import (
	// "time"
	// streamdeck "github.com/muesli/streamdeck"
	ui "github.com/0187773933/StreamDeck/v1/ui"
)

type ConfigFile struct {
	ServerBaseUrl string `yaml:"server_base_url"`
	ServerPort string `yaml:"server_port"`
	ServerAPIKey string `yaml:"server_api_key"`
	ServerCookieName string `yaml:"server_cookie_name"`
	ServerCookieSecret string `yaml:"server_cookie_secret"`
	ServerCookieAdminSecretMessage string `yaml:"server_cookie_admin_secret_message"`
	ServerCookieSecretMessage string `yaml:"server_cookie_secret_message"`
	ServerLoginURLPrefix string `yaml:"server_login_url_prefix"`
	AdminUsername string `yaml:"admin_username"`
	AdminPassword string `yaml:"admin_password"`
	TimeZone string `yaml:"time_zone"`
	BoltDBPath string `yaml:"bolt_db_path"`
	BoltDBEncryptionKey string `yaml:"bolt_db_encryption_key"`
	StreamDeckUI ui.StreamDeckUI `yaml:"stream_deck_ui"`
}