package main

//surveyCreate
//surveyVote
//surveyUpdate
//surveyEnd

//SurveyOption is
type SurveyOption struct {
	ID    int    `json:"id"`
	Text  string `json:"text"`
	Count int    `json:"count"`
}

//Survey is
type Survey struct {
	ID               int
	Text             string          `json:"text"`
	Options          []SurveyOption  `json:"options"`
	Owner            DataChannelUser `json:"owner"`
	ParticipantCount int
}
