package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/samplebuilder"
	"gopkg.in/hraban/opus.v2"
)

var peerConnectionConfig = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

const (
	rtcpPLIInterval = time.Second
	// mode for frames width per timestamp from a 30 second capture
	rtpAverageFrameWidth = 7
)

const SAMPLERATE = 48000
const CHANNELS = 2

//MediaRoom is
type MediaRoom struct {
	RoomID              int
	api                 *webrtc.API
	audioTrack          *webrtc.Track
	videoTrack          *webrtc.Track
	audioSampleBuilders []*samplebuilder.SampleBuilder
}

//NewMediaRoom is
func NewMediaRoom(id int) *MediaRoom {
	m := webrtc.MediaEngine{}

	codecAudio := webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, SAMPLERATE)
	codecVideo := webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000)

	m.RegisterCodec(codecAudio)
	m.RegisterCodec(codecVideo)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m))

	audioTrack, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeOpus, rand.Uint32(), "pion", "audio", codecAudio)
	if err != nil {
		panic(err)
	}

	videoTrack, err := webrtc.NewTrack(webrtc.DefaultPayloadTypeVP8, rand.Uint32(), "pion", "video", codecVideo)
	if err != nil {
		panic(err)
	}

	return &MediaRoom{
		RoomID:              id,
		api:                 api,
		audioTrack:          audioTrack,
		videoTrack:          videoTrack,
		audioSampleBuilders: []*samplebuilder.SampleBuilder{},
	}
}

//NewPeerConnection is
func (mediaRoom *MediaRoom) NewPeerConnection() *webrtc.PeerConnection {
	pc, err := mediaRoom.api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		panic(err)
	}
	return pc
}

//AddUser is
func (mediaRoom *MediaRoom) AddUser(
	user *User,
	room *Room,
	sd webrtc.SessionDescription,
	isPublisher bool) (*webrtc.SessionDescription, error) {
	pc := mediaRoom.NewPeerConnection()

	if isPublisher {
		_, err := pc.AddTrack(mediaRoom.audioTrack)
		if err != nil {
			return nil, err
		}

		// Allow us to receive 1 video track
		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}

		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

		// Set a handler for when a new remote track starts
		// Add the incoming track to the list of tracks maintained in the server
		addOnTrack(pc, mediaRoom.audioTrack, mediaRoom.videoTrack, mediaRoom)

		log.Println("Publisher")
	} else {
		_, err := pc.AddTrack(mediaRoom.audioTrack)
		if err != nil {
			return nil, err
		}
		_, err = pc.AddTrack(mediaRoom.videoTrack)
		if err != nil {
			return nil, err
		}

		if _, err := pc.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

		addOnTrackAudio(pc, mediaRoom.audioTrack, mediaRoom)

		log.Println("Client")
	}

	roomUser := &RoomUser{
		User:           user,
		PeerConnection: pc,
	}

	room.addRoomUser(roomUser)

	log.Printf("Added RoomUser id : %d. RoomUser Count : %d", user.ID, len(room.Users))

	DataChannelHandler(pc, room, roomUser)

	// Set the remote SessionDescription
	err := pc.SetRemoteDescription(sd)
	if err != nil {
		panic(err)
	}

	// Create answer
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		panic(err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = pc.SetLocalDescription(answer)
	if err != nil {
		panic(err)
	}

	return &answer, nil
}

//AddStreamer is
func (mediaRoom *MediaRoom) AddStreamer() {

}

//AddViewer is
func (mediaRoom *MediaRoom) AddViewer() {

}

//MediaRoomRepository is
type MediaRoomRepository struct {
	mediaRooms map[int]*MediaRoom
}

//NewMediaRoomRepository is
func NewMediaRoomRepository() *MediaRoomRepository {
	return &MediaRoomRepository{
		mediaRooms: make(map[int]*MediaRoom),
	}
}

var mediaRoomRepository *MediaRoomRepository = NewMediaRoomRepository()

//GetMediaRoomByRoomID is
func (repo *MediaRoomRepository) GetMediaRoomByRoomID(id int) *MediaRoom {
	mediaRoom := repo.mediaRooms[id]
	if mediaRoom == nil {
		mediaRoom = NewMediaRoom(id)
		sendAudioTrackToAllPeers(mediaRoom)
		repo.mediaRooms[id] = mediaRoom
	}
	return mediaRoom
}

//MessageType is
type MessageType string

const (
	chatControl = "chatControl"
	chatMessage = "chatMessage"
)

//ChatMessageData is
type ChatMessageData struct {
	UserID   string
	UserName string
	Message  string
}

//DataChannelHandler is
func DataChannelHandler(pc *webrtc.PeerConnection, room *Room, roomUser *RoomUser) {
	// Register data channel creation handling
	pc.OnDataChannel(func(d *webrtc.DataChannel) {
		roomUser.DataChannel = d

		log.Printf("New DataChannel %s - %d\n", d.Label(), d.ID())

		d.OnOpen(func() {
			log.Printf("Open Data channel %s - %d\n", d.Label(), d.ID())
		})

		d.OnClose(func() {
			log.Printf("Closed Data channel %s - %d.\n", d.Label(), d.ID())
		})

		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			var msg1 ChatMessage
			err := json.Unmarshal(msg.Data, &msg1)
			if err != nil {
				log.Fatalf("An error occured when parsing coming message %v", err)
			}
			log.Println(msg1)

			SendChatMessage(room, &msg1)
		})
	})
}

//ChatMessage is
type ChatMessage struct {
	Text string `json:"text"`
	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

//SendChatMessage is
func SendChatMessage(room *Room, chatMessage *ChatMessage) {
	user := userRepository.GetUserByID(chatMessage.User.ID)
	if user == nil {
		log.Fatalf("User not found! id : %d", chatMessage.User.ID)
		return
	}
	chatMessage.User.Name = user.Name

	room.addChatMessage(chatMessage)

	json, err := json.Marshal(chatMessage)
	if err != nil {
		log.Fatalf("An error occured at converting ChatMessage to json : %v", err)
		return
	}
	for _, roomUser := range room.Users {
		roomUser.DataChannel.Send(json)
	}
}

func addOnTrack(pc *webrtc.PeerConnection, audioTrack, videoTrack *webrtc.Track, mediaRoom *MediaRoom) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	pc.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}})
				if rtcpSendErr != nil {
					if rtcpSendErr == io.ErrClosedPipe {
						return
					}
					log.Println(rtcpSendErr)
				}
			}
		}()

		log.Println("Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		builderVideo := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.VP8Packet{})
		builderAudio := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.OpusPacket{})
		//builderAudio := audioBuilder
		mediaRoom.audioSampleBuilders = append(mediaRoom.audioSampleBuilders, builderAudio)

		for {
			rtp, err := remoteTrack.ReadRTP()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Panic(err)
			}

			if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
				builderAudio.Push(rtp)
				//for s := builderAudio.Pop(); s != nil; s = builderAudio.Pop() {
				//	if err := audioTrack.WriteSample(*s); err != nil && err != io.ErrClosedPipe {
				//		log.Panic(err)
				//	}
				//}
			} else if remoteTrack.Kind() == webrtc.RTPCodecTypeVideo {
				builderVideo.Push(rtp)
				for s := builderVideo.Pop(); s != nil; s = builderVideo.Pop() {
					if err := videoTrack.WriteSample(*s); err != nil && err != io.ErrClosedPipe {
						log.Panic(err)
					}
				}
			}
		}
	})
}

func addOnTrackAudio(pc *webrtc.PeerConnection, audioTrack *webrtc.Track, mediaRoom *MediaRoom) {
	// Set a handler for when a new remote track starts, this just distributes all our packets
	// to connected peers
	pc.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
		// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
		go func() {
			ticker := time.NewTicker(rtcpPLIInterval)
			for range ticker.C {
				rtcpSendErr := pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}})
				if rtcpSendErr != nil {
					if rtcpSendErr == io.ErrClosedPipe {
						return
					}
					log.Println(rtcpSendErr)
				}
			}
		}()

		log.Println("Audio Track acquired", remoteTrack.Kind(), remoteTrack.Codec())

		builderAudio := samplebuilder.New(rtpAverageFrameWidth*5, &codecs.OpusPacket{})
		//builderAudio := audioBuilder
		mediaRoom.audioSampleBuilders = append(mediaRoom.audioSampleBuilders, builderAudio)

		for {
			rtp, err := remoteTrack.ReadRTP()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Panic(err)
			}

			if remoteTrack.Kind() == webrtc.RTPCodecTypeAudio {
				builderAudio.Push(rtp)
				for s := builderAudio.Pop(); s != nil; s = builderAudio.Pop() {
					data := decode(s)
					encode(data, len(s.Data))
					//sample := media.Sample{Data: samplenew, Samples: s.Samples}
					if err := audioTrack.WriteSample(*s); err != nil && err != io.ErrClosedPipe {
						log.Panic(err)
					}
				}
			}
		}
	})
}

func sendAudioTrackToAllPeers(mediaRoom *MediaRoom) {
	return
	go func() {
		ticker := time.NewTicker(time.Millisecond * 10)
		for range ticker.C {
			if len(mediaRoom.audioSampleBuilders) <= 0 {
				continue
			}

			sampleLists := make([][]*media.Sample, len(mediaRoom.audioSampleBuilders))
			for i, audioBuilder := range mediaRoom.audioSampleBuilders {
				sampleList := sampleLists[i]
				for s := audioBuilder.Pop(); s != nil; s = audioBuilder.Pop() {
					sampleList = append(sampleList, s)
				}
				sampleLists[i] = sampleList
			}

			sort.SliceStable(sampleLists, func(i, j int) bool {
				return len(sampleLists[i]) > len(sampleLists[j])
			})

			for i := range sampleLists[0] {
				sumData, sampleCount, success := mixAllSamples(sampleLists, i)
				if success {
					if err := mediaRoom.audioTrack.WriteSample(media.Sample{Data: encode(sumData, 0), Samples: sampleCount}); err != nil && err != io.ErrClosedPipe {
						log.Panic(err)
					}
				}
			}
		}
	}()
}

func mixAllSamples(sampleLists [][]*media.Sample, i int) ([]int16, uint32, bool) {
	sumData := []int16{}
	var sampleCount uint32
	succes := false
	for j, sampleList := range sampleLists {
		if i < len(sampleList) {
			succes = true
			sample := sampleList[i]
			if j == 0 {
				sampleCount = sample.Samples
				sumData = decode(sample) // performance
			} else {
				//var err error
				//sumData, err = mix(sumData, decode(sample))
				//if err != nil {
				//	panic(err)
				//}
			}
		}
	}
	return sumData, sampleCount, succes
}

func decode(sample *media.Sample) []int16 {

	const G4 = 391.995
	const E3 = 164.814

	//const FRAME_SIZE_MS = 60
	//const FRAME_SIZE_MONO = SAMPLERATE * FRAME_SIZE_MS / 1000

	decoder, err := opus.NewDecoder(SAMPLERATE, CHANNELS)
	if err != nil {
		panic(err)
	}
	rawData := make([]int16, sample.Samples*CHANNELS)
	_, err = decoder.Decode(sample.Data, rawData)
	if err != nil {
		panic(err)
	}

	return rawData
}

func encode(rawData []int16, size int) []byte {

	const G4 = 391.995
	const E3 = 164.814
	//const FRAME_SIZE_MS = 60
	//const FRAME_SIZE_MONO = SAMPLERATE * FRAME_SIZE_MS / 1000

	encoder, err := opus.NewEncoder(SAMPLERATE, CHANNELS, opus.AppVoIP)
	if err != nil {
		panic(err)
	}
	data := make([]byte, size)
	_, err = encoder.Encode(rawData, data)
	if err != nil {
		panic(err)
	}

	return data
}

func byteToint16(data []byte) []int16 {
	result := make([]int16, len(data)/2)
	j := 0
	for i := 0; i < len(data); i += 2 {
		result[j] = int16(data[i])<<8 | int16(data[i+1])
		j++
	}
	return result
}

func mix(data1, data2 []int16) ([]int16, error) {
	if len(data1) != len(data2) {
		return nil, errors.New("data1 and data2 have different lenght")
	}
	sum := make([]int16, len(data1))
	for i := range data1 {
		sum[i] = data1[i] + data2[i]
	}

	return sum, nil
}
