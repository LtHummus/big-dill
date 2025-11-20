package socket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	dispatch *Dispatch
	conn     *websocket.Conn
	Send     chan []byte
	ID       string
}

func NewClient(dispatch *Dispatch, conn *websocket.Conn) {
	c := &Client{
		dispatch: dispatch,
		conn:     conn,
		Send:     make(chan []byte),
		ID:       uuid.NewString(),
	}

	go c.readPump()
	go c.WritePump()

	dispatch.RegistrationRequests <- c
}

func (c *Client) readPump() {
	log.Trace().Str("client_id", c.ID).Msg("starting read goroutine")
	defer func() {
		log.Debug().Str("client_id", c.ID).Msg("closing read goroutine")
		c.dispatch.DeregistrationRequests <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warn().Err(err).Str("client_id", c.ID).Msg("unexpected socket close")
			}
			break
		}

		var decodedmessage Message
		err = json.Unmarshal(message, &decodedmessage)
		if err != nil {
			log.Warn().Err(err).Str("client_id", c.ID).Msg("could not decode message")
			continue
		}
		c.dispatch.MessageInput <- &MessageWrapper{
			Source:  c,
			Message: &decodedmessage,
		}
	}
}

func (c *Client) WritePump() {
	log.Trace().Str("client_id", c.ID).Msg("starting write goroutine")

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Debug().Str("client_id", c.ID).Msg("closing write goroutine")
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// hub closed channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
