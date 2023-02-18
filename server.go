package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(nil)
	}

	uid := "sample"
	password := "hogehoge"

	username := "sample-user"
	misskeyPass := "fugafuga"
	if ldapLogin(uid, password) {
		createAccount(username, misskeyPass)
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
