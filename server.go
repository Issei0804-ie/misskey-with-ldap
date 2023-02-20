package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Issei0804-ie/misskey-with-ldap/auth"
	"github.com/Issei0804-ie/misskey-with-ldap/register"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	r := gin.Default()
	r.LoadHTMLGlob("templates/*/**")
	r.LoadHTMLGlob("templates/*.html")

	r.Static("/static", "./templates/static")
	r.GET("/", index)
	r.POST("/regist", regist)

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

func successRegister(c *gin.Context) {
	c.HTML(http.StatusAccepted, "register.html", gin.H{
		"title": "登録完了",
		"body":  "5秒後にmisskeyにリダイレクトします...",
	})
	return
}

func regist(c *gin.Context) {
	// TODO validate
	ldapUid := c.PostForm("ldap_username")
	ldapPassword := c.PostForm("ldap_password")
	misskeyUsername := c.PostForm("misskey_username")
	misskeyPassword := c.PostForm("misskey_password")

	var ldapUsernameError, ldapPasswordError, misskeyUsernameError, misskeyPasswordError string
	var isNotFill bool
	if ldapUid == "" {
		ldapUsernameError = "フィールドが空です"
		isNotFill = true
	}

	if ldapPassword == "" {
		ldapPasswordError = "フィールドが空です"
		isNotFill = true
	}
	if misskeyUsername == "" {
		misskeyUsernameError = "フィールドが空です"
		isNotFill = true
	}
	if misskeyPassword == "" {
		misskeyPasswordError = "フィールドが空です"
		isNotFill = true
	}

	if isNotFill {
		c.HTML(http.StatusBadRequest, "index.html", gin.H{
			"ldap_username_error":    ldapUsernameError,
			"ldap_password_error":    ldapPasswordError,
			"misskey_username_error": misskeyUsernameError,
			"misskey_password_error": misskeyPasswordError,
		})
		return
	}

	// TODO validate
	var l auth.Authenticator
	if os.Getenv("LDAP_MOCK") == "true" {
		l = auth.NewLDAPMock()
	} else {
		l = auth.NewLDAP(os.Getenv("LDAP_HOST"), os.Getenv("LDAP_MANAGER"), os.Getenv("LDAP_PASSWORD"), os.Getenv("LDAP_BASE"))
	}

	defer l.Close()
	if err := l.Login(ldapUid, ldapPassword); err != nil {
		log.Println(err)
		c.HTML(http.StatusBadRequest, "index.html", gin.H{
			"ldap_username_error": "LDAP認証が通りませんでした!!",
			"ldap_password_error": "LDAP認証が通りませんでした!!",
		})
		return
	}

	var instance register.Register
	if os.Getenv("MISSKEY_MOCK") == "true" {
		instance = register.NewMisskeyMockInstance()
	} else {
		instance = register.NewMisskeyInstance(os.Getenv("MISSKEY_HOST"), os.Getenv("MISSKEY_TOKEN"))
	}
	err := instance.SignUp(misskeyUsername, misskeyPassword)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "index.html", gin.H{
			"misskey_username_error": "misskeyへの登録が通りませんでした!! syskanに聞いて!!",
			"misskey_password_error": "misskeyへの登録が通りませんでした!! syskanに聞いて!!",
		})
		return
	}
	c.HTML(http.StatusAccepted, "register.html", gin.H{
		"title": "登録完了",
		"body":  "5秒後にmisskeyにリダイレクトします...",
	})
	return
}
