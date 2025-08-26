# GOOD FOR YOU


## TODO
- add questions for other domains, 
- logic to skip 'good' domains
    - if they are significantly better than the weakest?
    - do mental health always?
    - multiple questionnaire pages




## Endpoints

### GET /v1/userid
Generates and returns a user ID.

### GET /v1/questions/USERID/10
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