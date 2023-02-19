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
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
		return
	})

	r.POST("/register", func(c *gin.Context) {
		ldapUid := c.PostForm("ldap_username")
		ldapPassword := c.PostForm("ldap_password")
		misskeyUsername := c.PostForm("misskey_username")
		misskeyPassword := c.PostForm("misskey_password")
		if ldapLogin(ldapUid, ldapPassword) {
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
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"title": "LDAP認証失敗",
			"body":  "LDAP認証に失敗しました。パスワードとユーザー名が違うかもしれません。",
		})
	})
	if os.Getenv("SSL") == "true" {
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

func ldapLogin(uid string, password string) bool {

	ldapURI := os.Getenv("LDAP_HOST")
	l, err := ldap.DialURL(ldapURI)
	if err != nil {
		log.Fatal(nil)
	}
	defer l.Close()
	manager := os.Getenv("LDAP_MANAGER")
	managerPassword := os.Getenv("LDAP_PASSWORD")
	err = l.Bind(manager, managerPassword)
	if err != nil {
		log.Fatal(nil)
	}

	baseDN := os.Getenv("LDAP_BASE")

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(uid=%s))", uid),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	if len(sr.Entries) == 0 {
		log.Printf("user not found")
	} else if len(sr.Entries) != 1 {
		log.Println("to many found.")
	}

	entity := sr.Entries[0]
	err = l.Bind(entity.DN, password)

	if err != nil {
		log.Fatal(err)
	}
	return true
}

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
