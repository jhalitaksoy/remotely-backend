package main

import (
	"encoding/json"
	"log"
	"time"
)

type SurveyCreateParams struct {
	Text    string          `json:"text"`
	Options []*SurveyOption `json:"options"`
}

func OnSurveyCreate(room *Room, roomUser *Peer, data []byte) {
	var surveyCreateParams SurveyCreateParams
	err := json.Unmarshal(data, &surveyCreateParams)
	if err != nil {
		panic(err)
	}

	surveyFromUser := &Survey{
		Text:    surveyCreateParams.Text,
		Options: surveyCreateParams.Options,
	}

	survey := room.CreateNewSurvey(surveyFromUser)
	room.addSurvey(survey)

	sendNewSurveyToRoom(room, survey)
	log.Println("On New Survey")

	go func() {
		time.Sleep(surveyEndDuration)
		log.Println("On Survey Destroy! Survey ID : ", survey.ID)

		//surveyEndMessage := map[string]interface{}{ "surveyID": survey.ID}

		room.removeSurvey(survey)

		//sendMessageToRoom(room, surveyEnd, surveyEndMessage)
	}()
}

func sendNewSurveyToRoom(room *Room, survey *Survey) {
	//sendMessageToRoom(room, surveyCreate, survey)
}

func OnSurveyVote(room *Room, roomUser *Peer, data []byte) {

	var vote Vote
	err := json.Unmarshal(data, &vote)
	if err != nil {
		panic(err)
	}

	log.Printf("On Vote Survey. Room : %d, User : %d, Survey : %d, Vote : %d",
		room.ID, roomUser.User.ID, vote.SurveyID, vote.OptionID)

	survey := room.getSurveyByID(vote.SurveyID)

	if survey != nil {
		survey.Vote(roomUser.User, vote.OptionID)
		sendSurveyUpdatedToRoom(room, survey)
	} else {
		log.Printf("Survey Not Found! Room : %d, User : %d, Survey : %d, Vote : %d",
			room.ID, roomUser.User.ID, vote.SurveyID, vote.OptionID)
	}
}

func sendSurveyUpdatedToRoom(room *Room, survey *Survey) {
	//sendMessageToRoom(room, surveyUpdate, survey)
}
