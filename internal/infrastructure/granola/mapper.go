package granola

import (
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Mapper translates between Granola public API DTOs and domain types.
// This is the core of the anti-corruption layer — ensuring that
// external API concepts never leak into our domain model.

func mapNoteDetailToDomain(dto NoteDetailResponse) (*domain.Meeting, error) {
	participants := make([]domain.Participant, 0, len(dto.Attendees)+1)

	// Owner is the host
	if dto.Owner.Name != "" || dto.Owner.Email != "" {
		participants = append(participants, domain.NewParticipant(dto.Owner.Name, dto.Owner.Email, domain.RoleHost))
	}

	// Attendees (excluding owner to avoid duplicates)
	for _, a := range dto.Attendees {
		if a.Email == dto.Owner.Email && a.Name == dto.Owner.Name {
			continue
		}
		participants = append(participants, mapUserToDomain(a))
	}

	mtg, err := domain.New(
		domain.MeetingID(dto.ID),
		dto.Title,
		dto.CreatedAt,
		domain.SourceOther, // Public API doesn't expose source
		participants,
	)
	if err != nil {
		return nil, err
	}

	// Clear the creation event since this is a reconstitution, not a new creation
	mtg.ClearDomainEvents()

	// Map summary — prefer markdown if present, fall back to text
	summaryContent := dto.SummaryText
	summaryKind := domain.SummaryAuto
	if dto.SummaryMarkdown != nil && *dto.SummaryMarkdown != "" {
		summaryContent = *dto.SummaryMarkdown
	}
	if summaryContent != "" {
		mtg.AttachSummary(domain.NewSummary(domain.MeetingID(dto.ID), summaryContent, summaryKind))
		mtg.ClearDomainEvents()
	}

	// Clear all events — reconstitution should not produce events
	mtg.ClearDomainEvents()

	return mtg, nil
}

func mapNoteListItemToDomain(dto NoteListItem) (*domain.Meeting, error) {
	var participants []domain.Participant
	if dto.Owner.Name != "" || dto.Owner.Email != "" {
		participants = append(participants, domain.NewParticipant(dto.Owner.Name, dto.Owner.Email, domain.RoleHost))
	}

	mtg, err := domain.New(
		domain.MeetingID(dto.ID),
		dto.Title,
		dto.CreatedAt,
		domain.SourceOther,
		participants,
	)
	if err != nil {
		return nil, err
	}
	mtg.ClearDomainEvents()
	return mtg, nil
}

func mapUserToDomain(dto UserDTO) domain.Participant {
	return domain.NewParticipant(dto.Name, dto.Email, domain.RoleAttendee)
}

func mapTranscriptFromDetail(meetingID domain.MeetingID, dto NoteDetailResponse) *domain.Transcript {
	if len(dto.Transcript) == 0 {
		return nil
	}
	utterances := make([]domain.Utterance, len(dto.Transcript))
	for i, item := range dto.Transcript {
		utterances[i] = domain.NewUtterance(item.Speaker, item.Text, item.Timestamp, 0)
	}
	t := domain.NewTranscript(meetingID, utterances)
	return &t
}
