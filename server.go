package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/Issei0804-ie/misskey-with-ldap/auth"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", index)
	r.POST("/register", register)

	isSSL := os.Getenv("SSL") == "true"
	if isSSL {
		err := r.RunTLS(":443", "./ssl/server.pem", "./ssl/server.key")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := r.Run(":80")
		if err != nil {
			log.Fatal(err)
		}
	}
}

func index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
	return
}

func register(c *gin.Context) {
	// TODO validate
	ldapUid := c.PostForm("ldap_username")
	ldapPassword := c.PostForm("ldap_password")
	misskeyUsername := c.PostForm("misskey_username")
	misskeyPassword := c.PostForm("misskey_password")

	// TODO validate
	l := auth.NewLDAP(os.Getenv("LDAP_HOST"), os.Getenv("LDAP_MANAGER"), os.Getenv("LDAP_PASSWORD"), os.Getenv("LDAP_BASE"))
	defer l.Close()
	if err := l.Login(ldapUid, ldapPassword); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"title": "LDAP認証失敗",
			"body":  "LDAP認証に失敗しました。パスワードとユーザー名が違うかもしれません。",
		})
		return
	}
	instance := newMisskeyInstance(os.Getenv("MISSKEY_HOST"), os.Getenv("MISSKEY_TOKEN"))
	err := instance.SignUp(misskeyUsername, misskeyPassword)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "register.html", gin.H{
			"title": "登録失敗",
			"body":  err,
		})
		return
	}
	c.HTML(http.StatusAccepted, "register.html", gin.H{
		"title": "登録完了",
		"body":  "5秒後にmisskeyにリダイレクトします...",
	})
	return
}

type Register interface {
	SignUp(username, password string) error
}

func newMisskeyInstance(host, token string) Register {
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

	log.Println("%#v", string(byteArray))

	return nil
}
