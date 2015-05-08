package config

type Raven struct {
	Enable  bool   `json:"enable"`
	Address string `json:"address"`
}

type ConfigEntity struct {
	Raven Raven `json:"raven"`
}
