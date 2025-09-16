package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"user-db/db"
	"user-db/shared"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: admin <command> [arguments]")
		fmt.Println("\nAvailable Commands:")
		fmt.Println("  delete-questions           Delete all questions ")
		fmt.Println("  import-questions        Seed the database with initial question data")
		fmt.Println("  get-question <question-id>          Get question metadata")
		fmt.Println("  get <user-id>          Get questions and answers for the specified user")
		fmt.Println("  create-user <user-id>          create a new user with the specified user-id")
		fmt.Println("  delete-user <user-id>  delete user with specified user-id")
		fmt.Println("  add-answer <user-id>  <question-id> <value>         add an answer for user with question-id and value")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "get":
		getAnswersForUser(os.Args[2])
	case "create-user":
		createUser(os.Args[2])
	case "delete-user":
		deleteUser(os.Args[2])
	case "add-answer":
		addAnswer(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func createUser(userId string) {
	log.Printf("Creating new user with ID: %s", userId)
	db.NewUser(userId)
}

func deleteUser(userId string) {
	log.Printf("Deleting user with ID: %s", userId)
	db.DeleteUser(userId)
}

func addAnswer(args []string) {
	if len(args) != 3 {
		fmt.Println("Usage: admin add-answer <user-id> <question-id> <value>")
		os.Exit(1)
	}
	userId := args[0]
	questionId, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("Could not convert question-id to int: %s\n", args[1])
		os.Exit(1)
	}
	value, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("Could not convert value to int: %s\n", args[2])
		os.Exit(1)
	}
	log.Printf("Adding answer for user %s: question-id=%d, value=%d", userId, questionId, value)
	db.UpsertAnswer(userId, questionId, shared.SCALE, value)
}

func getAnswersForUser(userId string) {

	userAnswers, err := db.GetUser(userId)
	if err != nil {
		log.Printf("Error getting user (%s): %v", userId, err)
		os.Exit(1)
	}
	log.Printf("UserAnswers for user %s: %v", userId, userAnswers)
}
