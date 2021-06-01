package main

import "testing"

func TestEncodeDecodeMessage(t *testing.T) {
	channel := "channel"
	message := "message"
	data := EncodeMessage(MessageChannel(channel), Message(message))
	_channel, _message, _err := DecodeMessage(data)

	if _err != nil {
		t.Error(err)
	}

	if channel != string(_channel) {
		t.Error("channels not equal")
	}

	if message != string(_message) {
		t.Error("messages not equal")
	}
}

func TestEncodeDecodeMessageDifferentCharset(t *testing.T) {
	channel := "şçö"
	message := "şğçöü"
	data := EncodeMessage(MessageChannel(channel), Message(message))
	_channel, _message, _err := DecodeMessage(data)

	if _err != nil {
		t.Error(err)
	}

	if channel != string(_channel) {
		t.Error("channels not equal")
	}

	if message != string(_message) {
		t.Error("messages not equal")
	}
}
