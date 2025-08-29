package db

import (
	"context"
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

func ImportQuestions(questions map[int]shared.Question) {

	log.Println("Importing questions into MongoDB..." + questions[1].Text)

	var qs []shared.Question
	for _, v := range questions {
		qs = append(qs, v)
	}

	collection := client.Database(DATABASE_NAME).Collection(QUESTIONS)
	result, err := collection.InsertMany(context.TODO(), qs)
	if err != nil {
		panic(err)
	}
	log.Printf("Inserted %d documents: %v", len(result.InsertedIDs), result.InsertedIDs)
}

func DeleteQuestions() {
	collection := client.Database(DATABASE_NAME).Collection(QUESTIONS)
	result, err := collection.DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		panic(err)
	}
	log.Printf("Deleted %d documents", result.DeletedCount)
}

func GetQuestion(questionID int) shared.Question {
	collection := client.Database(DATABASE_NAME).Collection(QUESTIONS)
	var result shared.Question
	filter := map[string]int{"questionid": questionID}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		panic(err)
	}
	return result
}

func NewUser(userid string) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var userAnswers UserAnswers
	userAnswers.UserID = userid

	result, err := collection.InsertOne(context.TODO(), userAnswers)
	if err != nil {
		panic(err)
	}
	log.Printf("Inserted new user %s", result.InsertedID)
}

func GetUser(userID string) (UserAnswers, bool) {
	collection := client.Database(DATABASE_NAME).Collection(USERANSWERS)
	var result UserAnswers
	filter := map[string]string{"_id": userID}
	singleResult := collection.FindOne(context.TODO(), filter)
	err := singleResult.Decode(&result)
	if err == mongo.ErrNoDocuments {
		log.Printf("No user found with ID: %s", userID)
		return UserAnswers{}, false
	} else if err != nil {
		panic(err)
	}
	log.Printf("result: %v", result)
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

func GetQuestionForUser(questionID int) shared.Question {
	collection := client.Database(DATABASE_NAME).Collection(QUESTIONS)
	var result shared.Question
	filter := map[string]int{"questionid": questionID}
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		panic(err)
	}
	return result
}

func UpsertAnswer(userid string, questionID int, kind shared.AnswerKind, value int) error {

	answer := AnswerEvent{
		Kind:      kind.String(),
		Value:     &value,
		UpdatedAt: time.Now(),
	}

	latestPath := "answers." + strconv.Itoa(questionID) + ".latestAnswer"

	coll := client.Database(DATABASE_NAME).Collection(USERANSWERS)

	filter := bson.M{"_id": userid}
	update := bson.M{
		"$set": bson.M{
			latestPath: answer,
		}}

	result, err := coll.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Panic(err)
		return err
	}
	log.Printf("Matched %d documents and updated %d documents.", result.MatchedCount, result.ModifiedCount)
	return err
}
