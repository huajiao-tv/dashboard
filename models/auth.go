package models

import (
	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
	"github.com/youlu-cn/go-ldap/ldap"
)

func Login(user, password string) (*ginp.UserInfo, error) {
	// connect to server
	c, err := ldap.Dial("tcp", config.GlobalConfig.LDAP.Address)
	if err != nil {
		return nil, err
	}

	// search
	req := &ldap.SearchRequest{
		BaseDN:       config.GlobalConfig.LDAP.BaseDN,
		Scope:        ldap.ScopeWholeSubtree,
		DerefAliases: ldap.DerefAlways,
		SizeLimit:    1,
		TimeLimit:    10,
		TypesOnly:    true,
		Filter: &ldap.EqualityMatch{
			Attribute: "cn",
			Value:     []byte(user),
		},
	}
	res, err := c.Search(req)
	if err != nil {
		return nil, err
	}

	// bind
	if err := c.Bind(res[0].DN, []byte(password)); err != nil {
		return nil, err
	}

	return &ginp.UserInfo{
		Name:   string(res[0].Attributes["cn"][0]),
		Email:  string(res[0].Attributes["mail"][0]),
		Avatar: config.DefaultAvatar,
		Roles:  []string{""},
	}, nil
}
