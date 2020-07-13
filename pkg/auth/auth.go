package auth

import (
	"io/ioutil"
	"proxy/pkg/utils"
	"strings"
)

type Auth interface {
	Check(string)bool
}

type BasicAuth struct {
	userName string
	pass string
	data utils.ConcurrentMap
}

func (b BasicAuth) Check(string2 string) bool {
	panic("implement me")
}

func NewBasicAuth() BasicAuth {
	return BasicAuth{
		data: utils.NewConcurrentMap(),
	}
}

func (ba *BasicAuth) AddFromFile(file string) (n int, err error) {
	_content, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	userpassArr := strings.Split(strings.Replace(string(_content), "\r", "", -1), "\n")
	for _, userpass := range userpassArr {
		if strings.HasPrefix("#", userpass) {
			continue
		}
		u := strings.Split(strings.Trim(userpass, " "), ":")
		if len(u) == 2 {
			ba.data.Set(u[0], u[1])
			n++
		}
	}
	return
}

func (ba *BasicAuth) Add(userpassArr []string) (n int) {
	for _, userpass := range userpassArr {
		u := strings.Split(userpass, ":")
		if len(u) == 2 {
			ba.data.Set(u[0], u[1])
			n++
		}
	}
	return
}

func (ba *BasicAuth) Total() (n int) {
	n = ba.data.Count()
	return
}



