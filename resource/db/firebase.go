package db

import (
	"fmt"

	"github.com/94peter/pica/util"

	"golang.org/x/net/context"

	firebase "firebase.google.com/go"
	firebaseauth "firebase.google.com/go/auth"
	firebasedb "firebase.google.com/go/db"

	"google.golang.org/api/option"
)

const (
	ClaimState      = "state"
	ClaimPermission = "permission"
)

type firebaseDB struct {
	credentialsFile string
	c               string
	ctx             context.Context
	dburl           string

	client     *firebasedb.Client
	authClient *firebaseauth.Client
}

func (db *firebaseDB) connectDB() (*firebasedb.Client, error) {
	if db.client != nil {
		return db.client, nil
	}
	opt := option.WithCredentialsFile(db.credentialsFile)
	app, err := firebase.NewApp(db.ctx, &firebase.Config{
		DatabaseURL: db.dburl,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	db.client, err = app.Database(db.ctx)
	return db.client, err
}

func (db *firebaseDB) C(c string) InterDB {
	db.c = c
	return db
}

func (db *firebaseDB) Save(doc interDoc) error {
	client, err := db.connectDB()
	if err != nil {
		return err
	}
	ref := util.StrAppend(db.c, "/", doc.GetID())
	return client.NewRef(ref).Set(db.ctx, doc)
}

func (db *firebaseDB) GetByID(id string, doc interface{}) error {
	client, err := db.connectDB()
	if err != nil {
		return err
	}
	ref := util.StrAppend(db.c, "/", id)
	return client.NewRef(ref).Get(db.ctx, doc)
}

func (db *firebaseDB) connectAuth() (*firebaseauth.Client, error) {
	if db.authClient != nil {
		return db.authClient, nil
	}
	opt := option.WithCredentialsFile(db.credentialsFile)
	app, err := firebase.NewApp(db.ctx, &firebase.Config{
		DatabaseURL: db.dburl,
	}, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}
	db.authClient, err = app.Auth(db.ctx)
	return db.authClient, err
}

func (db *firebaseDB) CreateUser(phone, displayName, email, pwd, permission string) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}
	params := (&firebaseauth.UserToCreate{}).
		UID(phone).
		PhoneNumber(util.StrAppend("+886", phone[1:])).
		Email(email).
		Password(pwd).
		DisplayName(displayName).
		Disabled(false)
	_, err = client.CreateUser(db.ctx, params)
	return err
}

func (db *firebaseDB) SetUserDisable(uid string, disable bool) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}
	params := (&firebaseauth.UserToUpdate{}).
		Disabled(disable)
	_, err = client.UpdateUser(db.ctx, uid, params)
	return err
}

func (db *firebaseDB) DeleteUser(uid string) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}
	return client.DeleteUser(db.ctx, uid)
}

func (db *firebaseDB) ChangePwd(uid string, pwd string) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}
	params := (&firebaseauth.UserToUpdate{}).
		Password(pwd)
	_, err = client.UpdateUser(db.ctx, uid, params)
	return err
}

func (db *firebaseDB) UpdateState(uid string, state string) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}
	record, err := client.GetUser(db.ctx, uid)
	if err != nil {
		return nil
	}
	claim := record.CustomClaims
	claim[ClaimState] = state
	params := (&firebaseauth.UserToUpdate{}).CustomClaims(claim)
	_, err = client.UpdateUser(db.ctx, uid, params)
	return err
}

func (db *firebaseDB) UpdateUser(uid, display, permission string) error {
	client, err := db.connectAuth()
	if err != nil {
		return err
	}

	params := &firebaseauth.UserToUpdate{}
	if display != "" {
		params = params.DisplayName(display)
	}

	if permission != "" {
		record, err := client.GetUser(db.ctx, uid)
		if err != nil {
			return nil
		}
		claim := record.CustomClaims
		claim[ClaimPermission] = permission
		params = params.CustomClaims(claim)
	}

	_, err = client.UpdateUser(db.ctx, uid, params)
	return err
}

func (db *firebaseDB) VerifyToken(idToken string) (string, error) {
	client, err := db.connectAuth()
	if err != nil {
		return "", err
	}
	token, err := client.VerifyIDToken(db.ctx, idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}
