package config

type Raven struct {
	Enable  bool   `json:"enable"`
	Address string `json:"address"`
}

type GA struct {
	Enable bool   `json:"enable"`
	Id     string `json:"id"`
}

type Signup struct {
	Disable bool `json:"disable"`
}

type ConfigEntity struct {
	Raven  Raven  `json:"raven"`
	GA     GA     `json:"ga"`
	Signup Signup `json:"signup"`
}
