package auth

import (
	"fmt"
	"github.com/go-ldap/ldap/v3"
	"log"
	"runtime"
)

type Authenticator interface {
	Login(uid, password string) error
	Close()
}

func NewLDAP(URI, manager, password, dn string) Authenticator {
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

func NewLDAPMock() Authenticator {
	return LDAPMock{loginReturnValueMock: nil}
}

type LDAPMock struct {
	loginReturnValueMock error
}

func (l LDAPMock) Login(uid, password string) error {
	pt, _, _, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(pt).Name()
	log.Printf("func name is %s, uid(%s) was passed, return value is %d", funcName, uid, l.loginReturnValueMock)
	return l.loginReturnValueMock
}

func (l LDAPMock) Close() {
	return
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
