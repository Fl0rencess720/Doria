package models

import "github.com/gorilla/websocket"

const (
	CmdInit          = 1
	CmdAnswer        = 2
	CmdOffer         = 3
	CmdCandidate     = 4
	CmdInitResp      = 101
	CmdAnswerResp    = 102
	CmdOfferResp     = 103
	CmdCandidateResp = 104
	CmdRegisterResp  = 105
)

type Peer struct {
	ID   string
	Conn *websocket.Conn
}

type Message struct {
	Cmd     int    `json:"command"`
	Payload []byte `json:"payload"`
}

type Request struct {
	SourceID string `json:"source"`
	TargetID string `json:"target"`
	Body     []byte `json:"body"`
}

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type RegisterPeerRequest struct {
	PeerID string `json:"peer_id" binding:"required"`
}

type ForwardMessageRequest struct {
	SourceID string `json:"source_id" binding:"required"`
	TargetID string `json:"target_id" binding:"required"`
	Cmd      int    `json:"command" binding:"required"`
	Body     []byte `json:"body" binding:"required"`
}
