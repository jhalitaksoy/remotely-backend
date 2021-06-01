package main

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type RealtimeMessageTransportWebSocket struct {
	MessageChannels map[MessageChannel]OnMessage
	Peers           map[PeerID]*Peer
	onConnected     OnConnectionChange
	onDisConnected  OnConnectionChange
}

func NewRealtimeMessageTransportWebSocket() *RealtimeMessageTransportWebSocket {
	return &RealtimeMessageTransportWebSocket{
		MessageChannels: make(map[MessageChannel]func(*Peer, MessageChannel, Message)),
		Peers:           make(map[int]*Peer),
	}
}

func (rtmt *RealtimeMessageTransportWebSocket) ListenTestChannel() {
	rtmt.Listen("test", func(p *Peer, mc MessageChannel, m Message) {
		rtmt.Send(p.User.ID, "test", m)
	})
}

func (rtmt *RealtimeMessageTransportWebSocket) Send(peerID PeerID, channel MessageChannel, message Message) error {
	client, ok := myContext.Hub.clientsMap[peerID]
	if ok {
		data := EncodeMessage(channel, message)
		client.mu.Lock()
		defer client.mu.Unlock()
		return client.conn.WriteMessage(websocket.BinaryMessage, data)
	} else {
		return ErrPeerNotConnected
	}
}

func (rtmt *RealtimeMessageTransportWebSocket) Listen(messageChannel MessageChannel, onMessage OnMessage) error {
	_, ok := rtmt.MessageChannels[messageChannel]
	if ok {
		return ErrAlreadyListening
	}
	rtmt.MessageChannels[messageChannel] = onMessage
	return nil
}

func (rtmt *RealtimeMessageTransportWebSocket) OnConnected(onConnectionChange OnConnectionChange) {
	rtmt.onConnected = onConnectionChange
}

func (rtmt *RealtimeMessageTransportWebSocket) OnDisConnected(onConnectionChange OnConnectionChange) {
	rtmt.onDisConnected = onConnectionChange
}

func (rtmt *RealtimeMessageTransportWebSocket) OnWebsocketMessage(peer *Peer, msg []byte) {
	channel, message, err := DecodeMessage(msg)
	if err != nil {
		log.Printf("An error occured when decoding message : %v", err)
		return
	}

	callback, ok := rtmt.MessageChannels[channel]
	if !ok {
		log.Printf("Callback not found for channel : [ %s ]", channel)
		return
	}

	callback(peer, MessageChannel(channel), message)
}

//Hub ...
type Hub struct {
	clients map[*UserClientWS]bool

	clientsMap map[int]*UserClientWS

	Messages chan *ClientMessage

	register chan *UserClientWS

	unregister chan *UserClientWS
}

//NewHub ...
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*UserClientWS]bool),
		Messages:   make(chan *ClientMessage),
		clientsMap: make(map[int]*UserClientWS),
		register:   make(chan *UserClientWS),
		unregister: make(chan *UserClientWS),
	}
}

//Run ...
func (h *Hub) Run() {
	_rtmt, ok := myContext.RTMT.(*RealtimeMessageTransportWebSocket)
	var rtmt RealtimeMessageTransportWebSocket
	if ok {
		rtmt = *_rtmt
	} else {
		log.Println("rtmt is not RealtimeMessageTransportWebSocket")
	}
	for {
		select {
		case client := <-h.register:
			log.Println("WebSocket Register")
			h.clients[client] = true
			h.clientsMap[client.Peer.User.ID] = client
			if rtmt.onConnected != nil {
				rtmt.onConnected(client.Peer)
			}
		case client := <-h.unregister:
			log.Println("WebSocket UnRegister")
			if _, ok := h.clients[client]; ok {
				if rtmt.onDisConnected != nil {
					rtmt.onDisConnected(client.Peer)
				}
				delete(h.clientsMap, client.Peer.User.ID)
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
}

//HandleMessage ..
func (h *Hub) HandleMessage() {
	for {
		select {
		case message := <-h.Messages:
			//log.Println("(WebSocket) Received message")
			rtmt, ok := myContext.RTMT.(*RealtimeMessageTransportWebSocket)
			if ok {
				rtmt.OnWebsocketMessage(message.Client.Peer, message.Data)
			} else {
				log.Println("rtmt is not RealtimeMessageTransportWebSocket")
			}
		}
	}
}

//ClientMessage ...
type ClientMessage struct {
	Data   []byte
	Client *UserClientWS
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 100 * 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// UserClientWS is a middleman between the websocket connection and the hub.
type UserClientWS struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	Peer *Peer

	mu sync.Mutex
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *UserClientWS) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.Messages <- &ClientMessage{
			Data:   message,
			Client: c,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *UserClientWS) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			/*n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}*/

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
