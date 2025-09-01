# GOOD FOR YOU


## TODO
- add questions for other domains, 
- logic to skip 'good' domains
    - if they are significantly better than the weakest?
    - do mental health always?
    - multiple questionnaire pages

- THere must be an easier way to get all these averages and stuff

- use database for answers
- auth
    - use mongodb _id as uuid?

- is there a social connection bias in the prompt?


- make mongodb connection string a secret (godotenv?)
- implement history in database





## Endpoints

### GET /v1/userid
Generates and returns a user ID.

### GET /v1/questions/USERID
Gets the next x (in this case 10) questions in order of priority of user with USERID

### POST /v1/responses
expects answers to questions from a specified user
Body:
```json
{
    "userId": "eoXTTT9A",
    "answers": [
        {"questionId": 1, "value":3},
        {"questionId": 2, "value":9}
    ]
}
```

### GET v1/topics/llm/USERID/x
Get the content of topic with priority x.
(for now make an LLM call to get customized details about the topic)



## Wording

### Dimensions
The GF12 categories we call wellbeing dimensions here


### Facets
A wellbeing dimension, e.g. mental health, has multiple facets, e.g. basic psychological needs and emotional regulation




### ADMIN CLI
```
# Build
go build -o admin cmd/admin/main.go

# Run commands
./admin migrate         # Will ask for confirmation
```