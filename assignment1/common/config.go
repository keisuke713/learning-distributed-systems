package common

import (
	"encoding/json"
	"os"
)

type Config struct {
	ClientAddress        string
	NimServerAddress     string
	TracingServerAddress string
	Secret               []byte
	TracingIdentity      string
}

func ParseConfig(filepath string) *Config {
	f, err := os.Open(filepath)
	if err != nil {
		CheckErr(err, "reading config file")
	}

	jsonDecoder := json.NewDecoder(f)
	config := new(Config)
	if jsonDecoder.Decode(config); err != nil {
		CheckErr(err, "parsing config data")
	}
	return config
}
