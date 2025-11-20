package survey

import (
	"errors"
	"maps"
	"sync"

	"github.com/rs/zerolog/log"
)

var validVotes = map[string]struct{}{
	"-3": {},
	"-2": {},
	"-1": {},
	"0":  {},
	"1":  {},
	"2":  {},
	"3":  {},
}

var (
	ErrVoteClosed   = errors.New("vote is closed")
	ErrAlreadyVoted = errors.New("already voted")
	ErrInvalidVote  = errors.New("invalid vote")
)

type Surveyor struct {
	votes        map[string]int
	voteKeysUsed map[string]struct{}
	votesOpen    bool
	voteLock     *sync.Mutex

	onVoteStatusChange func(newStatus bool)
}

func NewSurveyor() *Surveyor {
	s := &Surveyor{
		votes:        map[string]int{},
		voteKeysUsed: map[string]struct{}{},
		votesOpen:    false,
		voteLock:     &sync.Mutex{},
	}

	s.ResetVotes()
	return s
}

func (s *Surveyor) SetChangeNotificationFunction(f func(bool)) {
	s.onVoteStatusChange = f
}

func (s *Surveyor) VoteStatus(voteKey string) string {
	s.voteLock.Lock()
	defer s.voteLock.Unlock()

	if !s.votesOpen {
		return "votes_closed"
	}

	_, voted := s.voteKeysUsed[voteKey]
	if voted {
		return "already_voted"
	}

	return "can_vote"
}

func (s *Surveyor) OpenVote() {
	if s.votesOpen {
		return
	}
	s.ResetVotes()
	s.votesOpen = true
	log.Info().Msg("opening vote")
	go s.onVoteStatusChange(true)
}

func (s *Surveyor) CloseVote() {
	if !s.votesOpen {
		return
	}
	s.votesOpen = false
	log.Info().Msg("closing vote")
	go s.onVoteStatusChange(false)
}

func (s *Surveyor) VotesOpen() bool {
	return s.votesOpen
}

func (s *Surveyor) ResetVotes() {
	s.voteLock.Lock()
	defer s.voteLock.Unlock()

	s.votes = map[string]int{}
	s.voteKeysUsed = map[string]struct{}{}
	s.votesOpen = false

	for k := range validVotes {
		s.votes[k] = 0
	}
	log.Info().Msg("resetting vote")
}

func (s *Surveyor) SubmitVote(vote string, voteKey string) error {
	s.voteLock.Lock()
	defer s.voteLock.Unlock()

	if !s.votesOpen {
		return ErrVoteClosed
	}

	if _, used := s.voteKeysUsed[voteKey]; used {
		log.Warn().Str("vote_key", voteKey).Msg("used vote key already")
		return ErrAlreadyVoted
	}

	if _, voteOK := validVotes[vote]; !voteOK {
		log.Warn().Str("vote_key", voteKey).Str("vote", vote).Msg("invalid vote")
		return ErrInvalidVote
	}

	s.votes[vote]++
	s.voteKeysUsed[voteKey] = struct{}{}
	log.Info().Str("vote_key", voteKey).Str("vote", vote).Msg("vote submitted")
	return nil
}

func (s *Surveyor) GetVotes() map[string]int {
	s.voteLock.Lock()
	defer s.voteLock.Unlock()

	return maps.Clone(s.votes)
}
