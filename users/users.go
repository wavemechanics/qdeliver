package users

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Users struct {
	Version  int       `json:"version"`
	Accounts []Account `json:"accounts"`
}

type Account struct {
	Owner    string `json:"owner"`
	Domain   string `json:"domain"`
	URL      string `json:"url"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Notify   bool   `json:"notify"`
}

func Load(path string) (*Users, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	var users Users
	if err = dec.Decode(&users); err != nil {
		return nil, err
	}
	return &users, nil
}

func (u *Users) Lookup(owner, domain string) (*Account, error) {
	for _, account := range u.Accounts {
		if account.Owner == owner && account.Domain == domain {
			return &account, nil
		}
	}
	return nil, os.ErrNotExist
}

func (u *Users) Save(path string) error {
	buf, err := json.MarshalIndent(u, "", "    ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, buf, 0666)
}
