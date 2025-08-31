package internal

import "log"

// todo: nested/organized structure of json settings file
type Settings struct {
	AppDir          string `json:"app_dir"`
	DBMigrationsDir string `json:"db_migrations_dir"`

	DataDir string `json:"data_dir"`
	JobsDir string `json:"jobs_dir"`

	WebPort int `json:"web_port"`

	GrblDevicePort               string `json:"grbl_device_port"`
	GrblDeviceFirmwareVersion    string `json:"grbl_firmware_version"`
	GrblDeviceFirmwareBinaryPath string `json:"grbl_firmware_binary_path"`
}

var (
	settings Settings
)

func LoadSetting() {
	log.Printf("load settings")
}

func GetSettings() Settings {
	return settings
}

//todo: accept absolute dirs (app and data) from env variables and rest keep in json file?
