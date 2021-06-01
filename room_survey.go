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

func OnSurveyCreate(myContext *MyContext, peer *Peer, data []byte) {
	var surveyCreateParams SurveyCreateParams
	err := json.Unmarshal(data, &surveyCreateParams)
	if err != nil {
		panic(err)
	}

	surveyFromUser := &Survey{
		Text:    surveyCreateParams.Text,
		Options: surveyCreateParams.Options,
	}

	survey := peer.Room.CreateNewSurvey(surveyFromUser)
	peer.Room.addSurvey(survey)

	sendNewSurveyToRoom(myContext, peer.Room, survey)
	log.Println("On New Survey")

	go func() {
		time.Sleep(surveyEndDuration)
		log.Println("On Survey Destroy! Survey ID : ", survey.ID)
		peer.Room.removeSurvey(survey)
		sendSurveyDestroyToRoom(myContext, peer.Room, survey)
	}()
}

func sendNewSurveyToRoom(myContext *MyContext, room *Room, survey *Survey) {
	surveyJson, err := json.Marshal(survey)
	if err != nil {
		log.Println(err)
	}
	for _, peer := range room.Users {
		myContext.RTMT.Send(peer.User.ID, ChannelSurveyCreate, surveyJson)
	}
}

func sendSurveyDestroyToRoom(myContext *MyContext, room *Room, survey *Survey) {
	surveyEndMessage := map[string]interface{}{"surveyID": survey.ID}
	messageJson, err := json.Marshal(surveyEndMessage)
	if err != nil {
		log.Println(err)
	}
	for _, peer := range room.Users {
		myContext.RTMT.Send(peer.User.ID, ChannelSurveyDestroy, messageJson)
	}
}

func OnSurveyVote(myContext *MyContext, peer *Peer, data []byte) {

	var vote Vote
	err := json.Unmarshal(data, &vote)
	if err != nil {
		panic(err)
	}

	log.Printf("On Vote Survey. Room : %d, User : %d, Survey : %d, Vote : %d",
		peer.Room.ID, peer.User.ID, vote.SurveyID, vote.OptionID)

	survey := peer.Room.getSurveyByID(vote.SurveyID)

	if survey != nil {
		survey.Vote(peer.User, vote.OptionID)
		sendSurveyUpdatedToRoom(peer.Room, survey)
	} else {
		log.Printf("Survey Not Found! Room : %d, User : %d, Survey : %d, Vote : %d",
			peer.Room.ID, peer.User.ID, vote.SurveyID, vote.OptionID)
	}
}

func sendSurveyUpdatedToRoom(room *Room, survey *Survey) {
	surveyJson, err := json.Marshal(survey)
	if err != nil {
		log.Println(err)
	}
	for _, peer := range room.Users {
		myContext.RTMT.Send(peer.User.ID, ChannelSurveyUpdate, surveyJson)
	}
}
