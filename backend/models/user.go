package models

type User struct {
	UserID      int64  `db:"user_id"`
	Username    string `db:"username"`
	Password    string `db:"password"`
	IsSiteAdmin bool   `db:"is_site_admin"`
}

type MePermissionsView struct {
	UserID      int64    `json:"user_id"`
	Username    string   `json:"username"`
	Roles       []string `json:"roles"`
	IsSiteAdmin bool     `json:"is_site_admin"`
}
