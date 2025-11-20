package socket

import (
	"encoding/json"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/lthummus/big-dill/internal/survey"
)

type Dispatch struct {
	clients  map[*Client]struct{}
	surveyor *survey.Surveyor

	RegistrationRequests   chan *Client
	DeregistrationRequests chan *Client
	MessageInput           chan *MessageWrapper
}

func NewDispatch(surveyor *survey.Surveyor) *Dispatch {
	d := &Dispatch{
		clients:                map[*Client]struct{}{},
		RegistrationRequests:   make(chan *Client),
		DeregistrationRequests: make(chan *Client),
		MessageInput:           make(chan *MessageWrapper),
		surveyor:               surveyor,
	}

	surveyor.SetChangeNotificationFunction(d.onVoteStatusChange)

	go d.run()

	return d
}

func (d *Dispatch) run() {
	log.Info().Msg("starting dispatch handler goroutine")
	for {
		select {
		case msg := <-d.MessageInput:
			d.handleMessage(msg)
		case client := <-d.DeregistrationRequests:
			if _, ok := d.clients[client]; ok {
				delete(d.clients, client)
				close(client.Send)
				log.Info().Str("client_id", client.ID).Int("num_clients", len(d.clients)).Msg("deregistering client")
			}
		case client := <-d.RegistrationRequests:
			d.clients[client] = struct{}{}
			successMessage := Message{
				Kind: "connect_success",
				Payload: map[string]any{
					"client_id": client.ID,
				},
			}
			log.Info().Str("client_id", client.ID).Int("num_clients", len(d.clients)).Msg("registering client")
			payload, _ := json.Marshal(successMessage)
			client.Send <- payload
		}
	}
}

func (d *Dispatch) onVoteStatusChange(newStatus bool) {
	msg := Message{
		Kind: "vote_status_change",
		Payload: map[string]any{
			"new_status": newStatus,
		},
	}

	payload, _ := json.Marshal(msg)

	for curr := range d.clients {
		curr.Send <- payload
	}
}

func (d *Dispatch) handleVote(msg *MessageWrapper) {
	vote, ok := msg.Message.Payload["vote"].(string)
	if !ok {
		log.Warn().Str("client_id", msg.Source.ID).Msg("no vote found")
		return
	}

	key, ok := msg.Message.Payload["vote_key"].(string)
	if !ok {
		log.Warn().Str("client_id", msg.Source.ID).Msg("no vote key found")
		return
	}

	err := d.surveyor.SubmitVote(vote, key)
	if err != nil {
		if errors.Is(err, survey.ErrInvalidVote) {
			log.Warn().Str("vote_key", key).Str("vote", vote).Msg("invalid vote")
			return
		}
		if errors.Is(err, survey.ErrVoteClosed) {
			// TODO: send vote closed status message?
			return
		}
		if !errors.Is(err, survey.ErrAlreadyVoted) {
			log.Error().Err(err).Str("vote_key", key).Str("vote", vote).Msg("unknown error")
			return
		}
	}

	msg.Source.Send <- []byte(`{"kind":"vote_success","payload":{}}`)
}

func (d *Dispatch) handleQueryVoteStatus(msg *MessageWrapper) {
	voteKey, ok := msg.Message.Payload["vote_key"].(string)
	if !ok {
		return
	}

	voteStatus := d.surveyor.VoteStatus(voteKey)
	response := map[string]any{
		"kind": "vote_status",
		"payload": map[string]any{
			"status": voteStatus,
		},
	}

	payload, _ := json.Marshal(response)
	msg.Source.Send <- payload
}

func (d *Dispatch) handleMessage(msg *MessageWrapper) {
	log.Trace().Str("message_kind", msg.Message.Kind).Str("source", msg.Source.ID).Msg("message received")

	switch msg.Message.Kind {
	case "vote":
		d.handleVote(msg)
	case "query_vote_status":
		d.handleQueryVoteStatus(msg)
	default:
		log.Warn().Str("kind", msg.Message.Kind).Str("client_id", msg.Source.ID).Msg("invalid message kind")
	}
}
