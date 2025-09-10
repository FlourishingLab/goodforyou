package db

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"
	"user-db/shared"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var client *mongo.Client
var DATABASE_NAME string = "goodforyou"
var USERANSWERS string = "useranswers"
var QUESTIONS string = "questions"

func init() {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	uri := os.Getenv("MONGODB_URI")
	log.Printf("Connecting to MongoDB at %s", uri)

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	// Create a new client and connect to the server
	var err error

	client, err = mongo.Connect(opts)
	if err != nil {
		panic(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
}

func NewUser(userid string) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var userAnswers UserAnswers
	userAnswers.UserID = userid
	userAnswers.Answers = make(map[int]QuestionAnswers)
	userAnswers.Insights = make(map[string]Insight)

	result, err := collection.InsertOne(context.TODO(), userAnswers)
	if err != nil {
		panic(err)
	}
	log.Printf("Inserted new user %s", result.InsertedID)
}

func GetUser(userID string) (UserAnswers, bool) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var result UserAnswers
	filter := map[string]string{"userid": userID}
	singleResult := collection.FindOne(context.TODO(), filter)
	err := singleResult.Decode(&result)
	if err == mongo.ErrNoDocuments {
		log.Printf("No user found with ID: %s", userID)
		return UserAnswers{}, false
	} else if err != nil {
		// TODO handle more gracefully
		panic(err)
	}
	return result, true
}

func DeleteUser(userID string) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	filter := map[string]string{"_id": userID}
	singleResult := collection.FindOneAndDelete(context.TODO(), filter)
	if singleResult.Err() != nil {
		panic(singleResult.Err())
	}
	log.Printf("deleted user: %v", singleResult.Decode(UserAnswers{}))
}

func UpsertAnswer(userid string, questionID int, kind shared.AnswerKind, value int) error {

	answer := AnswerEvent{
		Kind:      kind.String(),
		Value:     &value,
		UpdatedAt: time.Now(),
	}

	latestPath := "answers." + strconv.Itoa(questionID) + ".latestAnswer"

	coll := client.Database(DATABASE_NAME).Collection(USERANSWERS)

	filter := bson.M{"userid": userid}
	update := bson.M{
		"$set": bson.M{
			latestPath: answer,
		}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Panic(err)
		return err
	}
	log.Printf("UpcertAnswer: Matched %d documents and updated %d documents.", result.MatchedCount, result.ModifiedCount)
	return err
}

func UpcertInsight(userid string, insightsName, insightBlob string) error {

	insight := Insight{
		InsightJson: json.RawMessage(insightBlob),
	}

	insightsPath := "insights." + insightsName

	filter := bson.M{"userid": userid}
	update := bson.M{
		"$set": bson.M{
			insightsPath: insight,
		}}

	coll := client.Database(DATABASE_NAME).Collection(USERANSWERS)

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Panic(err)
		return err
	}
	log.Printf("UpcertInsight: Matched %d documents and updated %d documents.", result.MatchedCount, result.ModifiedCount)
	return err
}
