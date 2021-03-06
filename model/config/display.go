package config

type DisplayConfig struct {
	Enabled   bool   `yaml:"enabled" mapstructure:"enabled"`
	Width     int    `yaml:"width" mapstructure:"width"`
	Height    int    `yaml:"height" mapstructure:"height"`
	Rotation  uint8  `yaml:"rotation" mapstructure:"rotation"`
	FrameRate uint8  `yaml:"frame_rate" mapstructure:"frame_rate"`

	Bus      string `yaml:"bus" mapstructure:"bus"`
	DCPin    int    `yaml:"dc_pin" mapstructure:"dc_pin"`
	CSPin    int    `yaml:"cs_pin" mapstructure:"cs_pin"`
	ResetPin int    `yaml:"reset_pin" mapstructure:"reset_pin"`
	BusyPin  int    `yaml:"busy_pin" mapstructure:"busy_pin"`
}
