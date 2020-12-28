package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	_ "github.com/joho/godotenv/autoload"
)

type credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

var (
	words = []string{
		"paroxysme",
		"dichotomie",
		"belliqueux",
		"sinécure",
		"paradoxal",
		"obsolète",
		"insipide",
		"ersatz",
		"dystopie",
	}

	replies = []string{
		"Ouais, c'est pas faux.",
		"Ah c'est pas faux.",
		"C'est pas faux.",
		"C'est pas faux !",
		"Ah ouais, c'est pas faux.",
		"C'est... Non je connais pas c'mot là.",
		"Je connais pas c'mot là.",
	}

	cache map[int64]bool
)

func init() {
	rand.Seed(time.Now().UnixNano())

	var err error
	cache, err = loadCache()
	if err != nil {
		log.Fatal("Failed to load cache: ", err)
	}
}

func main() {
	client, err := getClient(&credentials{
		AccessToken:       os.Getenv("ACCESS_TOKEN"),
		AccessTokenSecret: os.Getenv("ACCESS_TOKEN_SECRET"),
		ConsumerKey:       os.Getenv("CONSUMER_KEY"),
		ConsumerSecret:    os.Getenv("CONSUMER_SECRET"),
	})
	if err != nil {
		log.Fatal("Failed to initialize Twitter client: ", err)
	}

	query := getRandomItem(words)
	log.Printf("Looking for tweets containing the word %q\n", query)
	search, _, err := client.Search.Tweets(&twitter.SearchTweetParams{
		Query:      query + " AND exclude:retweets",
		ResultType: "mixed", // mixed, popular, recent
		Lang:       "fr",
		Count:      100,
	})
	if err != nil {
		log.Fatal("Failed to search tweets: ", err)
	}

	log.Printf("%d matching tweets found\n", len(search.Statuses))
	mostPopularTweet := getMostPopularTweet(search.Statuses)
	if mostPopularTweet.ID != 0 {
		log.Printf("Replying to the tweet %d by @%s\n", mostPopularTweet.ID, mostPopularTweet.User.ScreenName)
		message := fmt.Sprintf("@%s %s", mostPopularTweet.User.ScreenName, getRandomItem(replies))
		tweet, response, err := client.Statuses.Update(message, &twitter.StatusUpdateParams{
			InReplyToStatusID: mostPopularTweet.ID,
		})
		if err != nil {
			log.Fatal("Failed to reply to tweet: ", err)
		} else if response.StatusCode != http.StatusOK {
			log.Fatal("Failed to reply to tweet: ", response.Status)
		}
		log.Printf("Created new tweet %d\n", tweet.ID)

		cache[mostPopularTweet.ID] = true
		if err := saveCache(cache); err != nil {
			log.Fatal("Failed to save cache: ", err)
		}
	}
}

func getClient(credentials *credentials) (*twitter.Client, error) {
	config := oauth1.NewConfig(credentials.ConsumerKey, credentials.ConsumerSecret)
	token := oauth1.NewToken(credentials.AccessToken, credentials.AccessTokenSecret)

	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	_, _, err := client.Accounts.VerifyCredentials(&twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getMostPopularTweet(tweets []twitter.Tweet) (mostPopularTweet twitter.Tweet) {
	for _, tweet := range tweets {
		if _, ok := cache[tweet.ID]; ok {
			continue
		}

		if tweet.RetweetCount > mostPopularTweet.RetweetCount {
			mostPopularTweet = tweet
		}
	}
	return mostPopularTweet
}

func getRandomItem(slice []string) string {
	return slice[rand.Intn(len(slice))]
}
