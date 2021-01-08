package main

import (
	"fmt"
	"time"
)

const (
	surveyEndDuration = 60 * time.Second
)

//SurveyOption is
type SurveyOption struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

//Survey is
type Survey struct {
	ID               int
	Text             string                `json:"text"`
	Options          []*SurveyOption       `json:"options"`
	Owner            DataChannelUser       `json:"owner"`
	ParticipantCount int                   `json:"participantCount"`
	Votes            map[int]*SurveyOption `json:"votes"` // user id and option id
}

//ConvertMapForUser ...
func (survey *Survey) ConvertMapForUser(user *User) map[string]interface{} {

	mapJSON := make(map[string]interface{})
	mapJSON["id"] = survey.ID
	mapJSON["text"] = survey.Text
	mapJSON["options"] = survey.Options
	mapJSON["owner"] = survey.Owner
	mapJSON["participantCount"] = survey.ParticipantCount

	if userVote := survey.Votes[user.ID]; userVote != nil {
		mapJSON["userVote"] = userVote
	}
	return mapJSON
}

//ConvertToMap ...
func (survey *Survey) ConvertToMap() map[string]interface{} {
	mapJSON := make(map[string]interface{})
	mapJSON["id"] = survey.ID
	mapJSON["text"] = survey.Text
	mapJSON["options"] = survey.Options
	mapJSON["owner"] = survey.Owner
	mapJSON["participantCount"] = survey.ParticipantCount
	mapJSON["votes"] = survey.convertVotesToMap()
	return mapJSON
}

func (survey *Survey) convertVotesToMap() map[int]interface{} {
	votes := make(map[int]interface{})
	for key, value := range survey.Votes {
		votes[key] = value.ID
	}
	return votes
}

//Vote ...
func (survey *Survey) Vote(user *User, optionID int) error {
	var option *SurveyOption
	for _, eachOption := range survey.Options {
		if optionID == eachOption.ID {
			option = eachOption
			break
		}
	}
	if option == nil {
		return fmt.Errorf("Option Not Found in Survey, Option ID : %d", optionID)
	}
	option.Count++
	survey.Votes[user.ID] = option
	survey.ParticipantCount++
	return nil
}

//Vote is
type Vote struct {
	SurveyID int `json:"surveyID"`
	OptionID int `json:"optionID"`
}
