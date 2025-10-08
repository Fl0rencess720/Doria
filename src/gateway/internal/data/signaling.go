package data

import (
	"sync"

	"github.com/Fl0rencess720/Doria/src/gateway/internal/biz"
	"github.com/Fl0rencess720/Doria/src/gateway/internal/models"
	"github.com/gorilla/websocket"
)

type signalingRepo struct {
	answerPeers map[string]*models.Peer
	offerPeers  map[string]*models.Peer
	mu          sync.RWMutex
}

func NewSignalingRepo() biz.SignalingRepo {
	return &signalingRepo{
		answerPeers: make(map[string]*models.Peer),
		offerPeers:  make(map[string]*models.Peer),
	}
}

func (r *signalingRepo) RegisterAnswerPeer(peerID string, conn *websocket.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.answerPeers[peerID] = &models.Peer{
		ID:   peerID,
		Conn: conn,
	}
	return nil
}

func (r *signalingRepo) UnregisterAnswerPeer(peerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.answerPeers, peerID)
	return nil
}

func (r *signalingRepo) RegisterOfferPeer(peerID string, conn *websocket.Conn) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.offerPeers[peerID] = &models.Peer{
		ID:   peerID,
		Conn: conn,
	}
	return nil
}

func (r *signalingRepo) UnregisterOfferPeer(peerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.offerPeers, peerID)
	return nil
}

func (r *signalingRepo) GetAnswerPeer(peerID string) (*models.Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.answerPeers[peerID]
	return peer, exists
}

func (r *signalingRepo) GetOfferPeer(peerID string) (*models.Peer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.offerPeers[peerID]
	return peer, exists
}