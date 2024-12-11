package web

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
)

/*
 * face for cookie
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro defines
const (
	CookieHashKey = "04DgUStWPtd0aJmiR+L2z5xwbpPr/hmH" //"the-big-and-secret-fash-key-here"//
	CookieBlockKey = "5oR820YTUcjpBFAYlgbte8jEL9o2oHI2" //"lot-secret-of-characters-big-too"//
	CookieExpireSeconds = 86400 //xxx seconds
)

//face info
type Cookie struct {
	hashKey, blockKey string
	secureCookie      *securecookie.SecureCookie `secure cookie instance`
	expireTime        int
}

//construct
func NewCookie() *Cookie {
	//self init
	this := &Cookie{
		hashKey: CookieHashKey,
		blockKey: CookieBlockKey,
		expireTime:CookieExpireSeconds,
	}

	//inter init
	this.initSecureCookie()
	return this
}

//set secure key
func (f *Cookie) SetSecureKey(hashKey, blockKey string) error {
	//check
	if hashKey == "" || blockKey == "" {
		return errors.New("invalid parameter")
	}
	f.initSecureCookie()
	return nil
}

//set expire time
func (f *Cookie) SetExpireTime(seconds int) {
	f.expireTime = seconds
}

//delete cookie
func (f *Cookie) DelCookie(name,
		domain string,
		c *gin.Context,
	) error {
	//check
	if name == "" || c == nil {
		return errors.New("invalid parameter")
	}
	//destroy cookie
	c.SetCookie(name, "", -1, "/", domain, false, true)
	return nil
}

//get cookie
func (f *Cookie) GetCookie(
		key string,
		c *gin.Context,
	) (string, error) {
	//check
	if key == "" || c == nil {
		return "", errors.New("invalid parameter")
	}
	//get original value
	orgVal, err := c.Cookie(key)
	if err != nil {
		return "", err
	}
	//try decode pass jwt
	//jwt, err := f.jwt.Decode(orgVal)
	//if err != nil {
	//	return nil, err
	//}
	//return jwt, nil
	return orgVal, nil
}

//set cookie
func (f *Cookie) SetCookie(
			key string,
			val string,
			expireSeconds int,
			domain string,
			c *gin.Context,
		) error {
	//check
	if key == "" || val == "" || c == nil {
		return errors.New("invalid parameter")
	}
	//set into cookie
	c.SetCookie(key, val, expireSeconds, "/", domain, false, true)
	return nil
}

//////////////
//private func
//////////////

func (f *Cookie) initSecureCookie() {
	//init security key
	hashKey := []byte(f.hashKey)
	blockKey := []byte(f.blockKey)
	//init cookie
	f.secureCookie = securecookie.New(hashKey, blockKey)
}