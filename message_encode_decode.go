package main

import "encoding/binary"

const headerLenght = 4

func EncodeMessage(messageChannel MessageChannel, message Message) []byte {
	channelLenght := len(messageChannel)
	messageLenght := len(message)

	totalLenght := headerLenght + channelLenght + messageLenght

	data := make([]byte, totalLenght)
	writeIntToByteArray(data[0:], uint32(channelLenght))
	writeStringToByteArray(data[headerLenght:], string(messageChannel))
	copy(data[headerLenght+channelLenght:], message)
	return data
}

func DecodeMessage(data []byte) (MessageChannel, Message, error) {
	channelLenght := readIntFromByteArray(data[0:])
	messageChannel := readStringFromByteArray(data[headerLenght : headerLenght+channelLenght])
	message := data[headerLenght+channelLenght:]
	return MessageChannel(messageChannel), Message(message), nil
}

func writeIntToByteArray(array []byte, value uint32) {
	binary.BigEndian.PutUint32(array, value)
}

func readIntFromByteArray(array []byte) uint32 {
	return binary.BigEndian.Uint32(array)
}

func writeStringToByteArray(array []byte, value string) {
	copy(array, []byte(value))
}

func readStringFromByteArray(array []byte) string {
	return string(array)
}
