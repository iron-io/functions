package models

type Config struct {
	DatabaseURL string `json:"db"`
	API         string `json:"api"`
	Logging     struct {
		To     string `json:"to"`
		Level  string `json:"level"`
		Prefix string `json:"prefix"`
	}
	PapertrailToken string `json:"papertrail_token"`
}

func (c *Config) Validate() error {
	return nil
}
