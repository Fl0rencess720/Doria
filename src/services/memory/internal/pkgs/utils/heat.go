package utils

import (
	"context"
	"math"
	"time"

	"github.com/Fl0rencess720/Doria/src/services/memory/internal/models"
)

const (
	HEAT_ALPHA        = 1.0
	HEAT_BETA         = 1.0
	HEAT_GAMMA        = 1.0
	RECENCY_TAU_HOURS = 24.0
)

func ComputeSegmentHeat(ctx context.Context, segment *models.Segment) (float64, error) {
	N_visit := float64(segment.Visit)

	L_interaction := float64(len(segment.Pages))

	R_recency := 1.0
	if !segment.LastVisit.IsZero() {
		R_recency = computeTimeDecay(segment.LastVisit, time.Now(), RECENCY_TAU_HOURS)
	}

	heat := HEAT_ALPHA*N_visit + HEAT_BETA*L_interaction + HEAT_GAMMA*R_recency

	return heat, nil
}

func computeTimeDecay(eventTime, currentTime time.Time, tauHours float64) float64 {
	deltaHours := currentTime.Sub(eventTime).Hours()
	return math.Exp(-deltaHours / tauHours)
}
