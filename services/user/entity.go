package user

type passwordEntity struct {
	Password string `json:"password"`
}

type userEntity struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email" valid:"email,required"`
	Password string `json:"password" valid:",required"`
	Admin    bool   `json:"admin" description:"isn't used now, set admin in config file"`
}
