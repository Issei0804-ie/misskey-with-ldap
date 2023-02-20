# miskey with LDAP

misskeyをLDAP経由で登録できるWEBアプリです。

## 立ち上げかた

sample.envをコピーして、.envを書きます。

```
cp sample.env .env
```

以下は記入例です。

```
MISSKEY_HOST="localhost"
MISSKEY_TOKEN="sample"
LDAP_HOST="localhost"
LDAP_PORT="389"
LDAP_PASSWORD="sample"
LDAP_BASE=""
LDAP_DN=""
#SSL=true
SSL=false
```


### Docker build

```
docker build -t issei0804-ie/misskey-with-ldap .
```

### docker run

```
docker run --rm  -p 443:443 -p 80:80 issei0804-ie/misskey-with-ldap:latest 
```