# GOOD FOR YOU


## TODO
- add questions for other domains, 

- Refactor. Clearer separation of concerns

- if a dimension has all responses, create a site for it in insights tab
  (- response endpoint signals if new insight (how to prevent wait?))
  - create list of domains with insights in frontend

- add uid and other labels to logs

- dimension http representation! "Meaning & Purpose" isn't good cause spaces and & character 

- use database for answers
- auth
    - use mongodb _id as uuid?

- api - have user-id as a variable for all/most calls (not in path)
- make mongodb connection string a secret (godotenv?)
- implement history in database





## Endpoints

### GET /v1/userid
Generates and returns a user ID.

### GET /v1/questions/<dimension>
Gets the next x (in this case 10) questions in order of priority of user with USERID

### POST /v1/responses
expects answers to questions from a specified user (in cookie)
side effect: if new responses lead to entirely answered dimension -> populate dimension insights
Body:
```json
{
    "answers": [
        {"questionId": 1, "value":3},
        {"questionId": 2, "value":9}
    ]
}
```


### GET v1/insights/llm/generate/holistic
executes holistic prompt

### GET v1/insights/llm
returns all insights for a user


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


### Deploy
Cloud Run is connected to the repo and is pulling, building and deploying new builds automatically