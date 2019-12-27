package dao

import (
	"errors"
	"strings"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
)

const (
	LdapUser = 1
)

var (
	UserNotExist      = errors.New("user not exist")
	InvalidCredential = errors.New("invalid username or password")
	Administration    = "administration"
)

type User struct {
	ginp.Model

	Name     string `binding:"required" json:"username" gorm:"unique_index"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Type     int    `json:"type"`
	Roles    string `json:"roles"`

	AuthSrc int `json:"-"`
}

func (m User) TableName() string {
	return "user"
}

func (m User) Create() error {
	return config.Postgres.Create(&m).Error
}

func (m User) Get() (v *User, err error) {
	db := config.Postgres.Model(&m).Where("name = ?", m.Name).First(&m)
	if db.RecordNotFound() {
		return nil, UserNotExist
	} else {
		return &m, db.Error
	}
}
func (m User) GetByID() (v *User, err error) {
	db := config.Postgres.Model(&m).Where("id = ?", m.ID).First(&m)
	if db.RecordNotFound() {
		return nil, UserNotExist
	} else {
		return &m, db.Error
	}
}

func CheckPermission(u interface{}, sys string) bool {
	var roles []string
	switch uEx := u.(type) {
	case *User:
		roles = strings.Split(uEx.Roles, ",")
	case *ginp.UserInfo:
		roles = uEx.Roles
	default:
		return false
	}
	for _, k := range roles {
		if k == Administration || k == sys {
			return true
		}
	}
	return false
}
func SearchUser(keyword string) (v []*User, err error) {
	db := config.Postgres.Model(&User{}).Where("name like ?", "%"+keyword+"%").Find(&v)
	for _, vv := range v {
		vv.Password = ""
	}
	return v, db.Error
}

func (m User) Delete() (err error) {
	return config.Postgres.Delete(&m).Error
}

func (m User) Update() error {
	return config.Postgres.Save(&m).Error
}
