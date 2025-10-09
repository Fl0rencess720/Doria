package utils

import (
	"fmt"

	"github.com/google/uuid"
)

func GenerateUniqueID() string {
	return uuid.New().String()[:8]
}

func GenerateOfferPeerID(sessionID string) string {
	if sessionID == "" {
		sessionID = GenerateUniqueID()
	}
	return fmt.Sprintf("offer-peer-%s", sessionID)
}

func GenerateAnswerPeerID(sessionID string) string {
	if sessionID == "" {
		sessionID = GenerateUniqueID()
	}
	return fmt.Sprintf("answer-peer-%s", sessionID)
}
