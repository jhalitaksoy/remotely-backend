package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSurveyJson(t *testing.T) {
	user := &User{
		ID:       0,
		Name:     "hlt",
		Password: "1234",
	}

	surveyOption1 := &SurveyOption{
		ID:    0,
		Text:  "optino1",
		Count: 0,
	}

	survey := Survey{
		Text:    "",
		Options: []*SurveyOption{surveyOption1},
		Votes: map[int]*SurveyOption{
			user.ID: surveyOption1,
		},
	}

	survey.Owner = DataChannelUser{
		ID:   user.ID,
		Name: user.Name,
	}
	bytes, err := json.Marshal(survey.ConvertToMap())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}
