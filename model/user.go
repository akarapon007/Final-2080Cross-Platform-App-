package model

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique"`
	Password string `json:"-"`
	Name     string `json:"name"`
	Address  string `json:"address"`
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == password
}
