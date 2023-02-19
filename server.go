package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_, pwd, _, _ := runtime.Caller(0)
	dir := filepath.Dir(pwd)
	log.Println(dir)
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(nil)
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
	ldapUid := c.PostForm("ldap_username")
	ldapPassword := c.PostForm("ldap_password")
	misskeyUsername := c.PostForm("misskey_username")
	misskeyPassword := c.PostForm("misskey_password")

	l := newLDAP(os.Getenv("LDAP_HOST"), os.Getenv("LDAP_MANAGER"), os.Getenv("LDAP_PASSWORD"), os.Getenv("LDAP_BASE"))
	defer l.Close()
	if err := l.Login(ldapUid, ldapPassword); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"title": "LDAP認証失敗",
			"body":  "LDAP認証に失敗しました。パスワードとユーザー名が違うかもしれません。",
		})
		return
	}
	err := createAccount(misskeyUsername, misskeyPassword)
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

type Authenticator interface {
	Login(uid, password string) error
	Close()
}

type Register interface {
	SignUp(username, password string) error
}

func newLDAP(URI, manager, password, dn string) Authenticator {
	l, err := ldap.DialURL(URI)
	if err != nil {
		log.Fatal(err)
	}
	return LDAP{
		URI:      URI,
		Manager:  manager,
		Password: password,
		conn:     l,
		DN:       dn,
	}
}

type LDAP struct {
	URI      string
	Manager  string
	Password string
	conn     *ldap.Conn
	DN       string
}

func (l LDAP) Close() {
	l.conn.Close()
	return
}

func (l LDAP) Login(uid, password string) error {
	err := l.conn.Bind(l.Manager, l.Password)
	if err != nil {
		log.Println(err)
		return err
	}

	searchRequest := ldap.NewSearchRequest(
		l.DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", uid),
		[]string{"dn"},
		nil,
	)

	sr, err := l.conn.Search(searchRequest)
	if err != nil {
		log.Println(err)
		return err
	}

	if len(sr.Entries) == 0 {
		log.Printf("user not found")
		return err
	} else if len(sr.Entries) != 1 {
		log.Println("to many found.")
		return err
	}

	entity := sr.Entries[0]
	err = l.conn.Bind(entity.DN, password)

	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// #TODO create mock
func createAccount(username, password string) error {
	endpoint, err := url.JoinPath(os.Getenv("MISSKEY_HOST"), "/api/admin/accounts/create")
	if err != nil {
		log.Println(err)
		return err
	}
	token := os.Getenv("MISSKEY_TOKEN")
	reqBody := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		I        string `json:"i"`
	}{
		Username: username,
		Password: password,
		I:        token,
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
