package util

import (
	"crypto/rsa"
	"errors"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtConf struct {
	PrivateKeyFile string `yaml:"privatekey"`
	PublicKeyFile  string `yaml:"publickey"`
	Header         struct {
		Alg string `yaml:"alg"`
		Typ string `yaml:"typ"`
		Kid string `yaml:"kid"`
	} `yaml:"header"`
	Claims struct {
		Iss         string        `yaml:"iss"`
		ExpDuration time.Duration `yaml:"exp"`
	} `yaml:"claims"`

	myHeader   map[string]interface{}
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

func (j *JwtConf) getHeader() map[string]interface{} {
	if j.myHeader != nil {
		return j.myHeader
	}
	j.myHeader = map[string]interface{}{
		"alg": j.Header.Alg,
		"typ": j.Header.Typ,
		"kid": j.Header.Kid,
	}
	return j.myHeader
}

func (j *JwtConf) getPublicKey() (*rsa.PublicKey, error) {
	if j.publicKey != nil {
		return j.publicKey, nil
	}
	publicData, err := ioutil.ReadFile(j.PublicKeyFile)
	if err != nil {
		return nil, err
	}
	j.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicData)
	return j.publicKey, err
}

func (j *JwtConf) getPrivateKey() (*rsa.PrivateKey, error) {
	if j.privateKey != nil {
		return j.privateKey, nil
	}
	privateData, err := ioutil.ReadFile(j.PrivateKeyFile)
	if err != nil {
		return nil, err
	}
	j.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateData)
	return j.privateKey, err
}

func (j *JwtConf) Parse(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		pk, err := j.getPublicKey()
		return pk, err
	})
	if token.Valid {
		return token, nil
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			return nil, errors.New("That's not even a token")
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			return nil, errors.New("Timing is everything")
		} else {
			return nil, err
		}
	}
	return nil, err

}

func (j *JwtConf) GetToken(data map[string]interface{}) (*string, error) {
	if data == nil {
		return nil, errors.New("no data")
	}
	now := time.Now()
	const expDuration = 3 * time.Hour
	data["iss"] = j.Claims.Iss
	data["iat"] = now.Unix()
	data["exp"] = now.Add(j.Claims.ExpDuration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims(data))

	token.Header = j.getHeader()

	pk, err := j.getPrivateKey()
	if err != nil {
		return nil, err
	}
	ss, err := token.SignedString(pk)
	if err != nil {
		return nil, err
	}
	return &ss, nil
}
