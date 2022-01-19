package db

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

//db variables
var Ctx context.Context
var Cancel context.CancelFunc
var Client *mongo.Client

var startOnce sync.Once
var endOnce sync.Once
var salt = "flexseal"

func StartDB(host string) error {
	//usage shown from https://github.com/mongodb/mongo-go-driver
	Ctx, Cancel = context.WithTimeout(context.Background(), 10*time.Second)
	var err error
	defer Cancel()
	Client, err = mongo.Connect(Ctx, options.Client().ApplyURI("mongodb://"+host+":27017"))
	if err != nil {
		log.Println(err)
	}
	log.Println("connected to DB")
	//create db for users
	if err = Client.Database("db").CreateCollection(Ctx, "users", nil); err != nil {
		log.Println(err)
	}
	//create db for tokens
	if err = Client.Database("db").CreateCollection(Ctx, "tokens", nil); err != nil {
		log.Println(err)
	}

	return nil
}

//return (true, <username>) if token is valid
//returns (false, <"">) if token is not valid
func IsValidToken(token string) (bool, string) {
	//return false, ""
	if token == "nil" {
		return false, ""
	}
	hashedToken := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	users := Client.Database("db").Collection("users")
	var databaseEntry bson.M
	//check for valid token
	if err := users.FindOne(context.Background(), bson.M{"token": string(hashedToken)}).Decode(&databaseEntry); err == nil {
		return true, databaseEntry["username"].(string)
	}
	return false, ""
}

//returns json representation of user data if user is in database
//returns nil if user wasn't found
func GetUserInfo(username string) []byte {
	var user bson.M
	users := Client.Database("db").Collection("users")
	if err := users.FindOne(context.Background(), bson.M{"username": username}).Decode(&user); err == nil {
		userJSON, _ := json.Marshal(user) //TODO test this
		return userJSON
	}
	return nil
}

//returns true if username is valid AND salted+hashed password == salt+hashed password in db
//returns (true, <token> if credentials are valid)
//returns (false, <"">) if credentials are invalid
func VerifyCredentials(username, password string) bool {
	password = salt + password

	collection := Client.Database("db").Collection("users")
	var databaseEntry bson.M

	if err := collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&databaseEntry); err == nil {
		if bcrypt.CompareHashAndPassword([]byte(string(databaseEntry["password"].(string))), []byte(password)) == nil {
			return true
		}
	}
	return false
}

//return true if user was registered
func RegisterUser(username, password string) bool {
	
	password = salt + password

	//store in database
	bytesPassword, _ := bcrypt.GenerateFromPassword([]byte(password), 0)

	//no duplicate usernames
	collection := Client.Database("db").Collection("users")
	var databaseEntry bson.M
	if collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&databaseEntry) == nil {
		return false
	}
	//user does not exist
	if _, err := collection.InsertOne(context.Background(), bson.M{"username": username, "password": string(bytesPassword), "token": "nil", "profilePic": "/"}); err != nil {
		log.Panic(err)
		return false
	}

	//registering won't authenticate user! they still need to log in
	return true
}

//simply stores a token for a given username
//should be called after a user logs in
func StoreToken(username, token string) error {
	collection := Client.Database("db").Collection("users")
	hashedToken := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	_, err := collection.UpdateOne(context.Background(), bson.M{"username": username}, bson.M{"$set": bson.M{"token": hashedToken}})
	if err != nil {
		log.Panic(err)
	}

	return nil
}

//retrieves the profile picture for a particular user
func GetProfilePath(username string) string {
	collection := Client.Database("db").Collection("users")
	var databaseEntry bson.M
	if collection.FindOne(context.Background(), bson.M{"username": username}).Decode(&databaseEntry) == nil { //no error
		return databaseEntry["profilePic"].(string)
	}
	return ""
}

//retrieves the profile picture for a particular user
func StoreProfilePath(username, filePath string) error {
	collection := Client.Database("db").Collection("users")
	_, err := collection.UpdateOne(context.Background(), bson.M{"username": username}, bson.M{"$set": bson.M{"profilePic": filePath}})
	if err != nil {
		log.Panic(err)
	}

	return nil
}

func closeDB() error {
	endOnce.Do(func() {
		if err := Client.Disconnect(Ctx); err != nil {
			panic(err)
		}
	})
	return nil
}
