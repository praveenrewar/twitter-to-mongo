package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ChimeraCoder/anaconda"
	mgo "gopkg.in/mgo.v2"
)

var accessToken string
var accessTokenSecret string
var consumerKey string
var consumerSecret string
var mongoHost string
var mongoPort int
var database = "Twitter"
var collection = "tweets"

func init() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var config Config
	json.Unmarshal(byteValue, &config)
	accessToken = config.AccessToken
	accessTokenSecret = config.AccessTokenSecret
	consumerKey = config.ConsumerKey
	consumerSecret = config.ConsumerSecret
	mongoHost = config.MongoHost
	mongoPort = config.MongoPort
}

func main() {
	session, err := mgo.Dial(mongoHost + ":" + strconv.Itoa(mongoPort))
	if err != nil {
		panic(err)
	}
	c := session.DB(database).C(collection)
	api := anaconda.NewTwitterApiWithCredentials(accessToken, accessTokenSecret, consumerKey, consumerSecret)
	v := url.Values{}
	s := api.PublicStreamSample(v)
	var count = 0
	go func() {
		for {
			fmt.Println("Tweets Inserted in Mongo: ", count)
			time.Sleep(5 * time.Second)
		}
	}()
	for t := range s.C {
		// fmt.Println("Here", t)
		var tweet Tweet
		switch v := t.(type) {
		case anaconda.Tweet:
			tweet.UserName = v.User.Name
			tweet.UserHandle = v.User.ScreenName
			tweet.CreatedAt = v.CreatedAt
			tweet.Text = v.FullText
			var hashtags []string
			for _, value := range v.Entities.Hashtags {
				hashtags = append(hashtags, value.Text)
			}
			tweet.Hashtags = hashtags
			// tweetJSON, err := json.Marshal(tweet)
			// if err != nil {
			// 	fmt.Println("Error Marshling json: ", err.Error())
			// 	continue
			// }
			err = c.Insert(tweet)
			if err != nil {
				log.Println("Error in inserting data to: ", err.Error())
			}
			count++
		}
	}
}

//Tweet is used to store data relevant data of a tweet
type Tweet struct {
	UserName   string   `json:"name"`
	UserHandle string   `json:"handle"`
	CreatedAt  string   `json:"created_at"`
	Text       string   `json:"text"`
	Hashtags   []string `json:"hashtags"`
}

//Config is used to store various config details
type Config struct {
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	MongoHost         string `json:"mongo_host"`
	MongoPort         int    `json:"mongo_port"`
}
