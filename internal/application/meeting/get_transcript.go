package meeting

import (
	"context"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

type GetTranscriptInput struct {
	MeetingID domain.MeetingID
}

type GetTranscriptOutput struct {
	Transcript *domain.Transcript
}

type GetTranscript struct {
	repo domain.Repository
}

func NewGetTranscript(repo domain.Repository) *GetTranscript {
	return &GetTranscript{repo: repo}
}

func (uc *GetTranscript) Execute(ctx context.Context, input GetTranscriptInput) (*GetTranscriptOutput, error) {
	if input.MeetingID == "" {
		return nil, domain.ErrInvalidMeetingID
	}

	t, err := uc.repo.GetTranscript(ctx, input.MeetingID)
	if err != nil {
		return nil, err
	}

	return &GetTranscriptOutput{Transcript: t}, nil
}
