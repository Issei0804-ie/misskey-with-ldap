package register

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Register interface {
	SignUp(username, password string) error
}

func NewMisskeyMockInstance() Register {
	return MissKeyMockInstance{}
}

type MissKeyMockInstance struct {
}

func (m MissKeyMockInstance) SignUp(username string, password string) error {
	return nil
}

func NewMisskeyInstance(host, token string) Register {
	return MissKeyInstance{
		host:  host,
		token: token,
	}
}

type MissKeyInstance struct {
	host  string
	token string
}

func (m MissKeyInstance) SignUp(username, password string) error {
	endpoint, err := url.JoinPath(m.host, "/api/admin/accounts/create")
	if err != nil {
		log.Println(err)
		return err
	}
	reqBody := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		I        string `json:"i"`
	}{
		Username: username,
		Password: password,
		I:        m.token,
	}
	jsonString, err := json.Marshal(reqBody)

	if err != nil {
		log.Println(err)
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(jsonString))
	if err != nil {
		log.Println(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println(resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return errors.New("登録エラー")
	}
	log.Println("%#v", string(byteArray))

	return nil
}
