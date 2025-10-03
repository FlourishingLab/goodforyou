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
		log.Fatalf("Error connecting to MongoDB: %s", err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatalf("Error pinging MongoDB: %s", err)
	}
}

func NewUser(userid string) error {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var userAnswers UserAnswers
	userAnswers.UserID = userid
	userAnswers.Answers = make(map[int]QuestionAnswers)
	userAnswers.Insights = make(map[string]Insight)
	userAnswers.Paragraphs = make(map[int]Paragraph)

	_, err := collection.InsertOne(context.TODO(), userAnswers)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(userID string) (UserAnswers, error) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var result UserAnswers
	filter := map[string]string{"userid": userID}
	singleResult := collection.FindOne(context.TODO(), filter)
	err := singleResult.Decode(&result)

	return result, err
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

	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func UpsertInsight(userid string, insightsName, insightBlob string, status InsightStatus) error {

	insightsPath := "insights." + insightsName

	filter := bson.M{"userid": userid}

	var update bson.M
	insight := Insight{
		Status: status,
	}
	if status == GENERATING {
		update = bson.M{
			"$set": bson.M{
				insightsPath: insight,
			}}
	} else {
		insight.InsightJson = json.RawMessage(insightBlob)
		update = bson.M{
			"$set": bson.M{
				insightsPath: insight,
			}}
	}

	coll := client.Database(DATABASE_NAME).Collection(USERANSWERS)

	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}

func NewDay(userid string, streakNumber int, paragraphNumber int) error {

	paragraphPath := "paragraph." + strconv.Itoa(paragraphNumber) + ".wasShown"

	filter := bson.M{"userid": userid}

	update := bson.M{
		"$set": bson.M{
			paragraphPath: true,
			"streak":      streakNumber,
			"lastVisited": time.Now(),
		}}

	coll := client.Database(DATABASE_NAME).Collection(USERANSWERS)

	_, err := coll.UpdateOne(context.TODO(), filter, update)

	return err
}
