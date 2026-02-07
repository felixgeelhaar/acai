package granola

import (
	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Mapper translates between Granola API DTOs and domain types.
// This is the core of the anti-corruption layer — ensuring that
// external API concepts never leak into our domain model.

func mapDocumentToDomain(dto DocumentDTO) (*domain.Meeting, error) {
	participants := make([]domain.Participant, len(dto.Participants))
	for i, p := range dto.Participants {
		participants[i] = mapParticipantToDomain(p)
	}

	mtg, err := domain.New(
		domain.MeetingID(dto.ID),
		dto.Title,
		dto.CreatedAt,
		mapSourceToDomain(dto.Source),
		participants,
	)
	if err != nil {
		return nil, err
	}

	// Clear the creation event since this is a reconstitution, not a new creation
	mtg.ClearDomainEvents()

	if dto.Summary != nil {
		mtg.AttachSummary(mapSummaryToDomain(domain.MeetingID(dto.ID), *dto.Summary))
		mtg.ClearDomainEvents()
	}

	for _, ai := range dto.ActionItems {
		item, err := mapActionItemToDomain(domain.MeetingID(dto.ID), ai)
		if err != nil {
			continue
		}
		mtg.AddActionItem(item)
	}

	if dto.Metadata != nil {
		mtg.SetMetadata(mapMetadataToDomain(*dto.Metadata))
	}

	// Clear all events — reconstitution should not produce events
	mtg.ClearDomainEvents()

	return mtg, nil
}

func mapParticipantToDomain(dto ParticipantDTO) domain.Participant {
	role := domain.RoleAttendee
	if dto.Role == "host" {
		role = domain.RoleHost
	}
	return domain.NewParticipant(dto.Name, dto.Email, role)
}

func mapSourceToDomain(source string) domain.Source {
	switch source {
	case "zoom":
		return domain.SourceZoom
	case "google_meet":
		return domain.SourceMeet
	case "teams":
		return domain.SourceTeams
	default:
		return domain.SourceOther
	}
}

func mapSummaryToDomain(meetingID domain.MeetingID, dto SummaryDTO) domain.Summary {
	kind := domain.SummaryAuto
	if dto.Type == "user_edited" {
		kind = domain.SummaryEdited
	}
	return domain.NewSummary(meetingID, dto.Content, kind)
}

func mapActionItemToDomain(meetingID domain.MeetingID, dto ActionItemDTO) (*domain.ActionItem, error) {
	item, err := domain.NewActionItem(
		domain.ActionItemID(dto.ID),
		meetingID,
		dto.Owner,
		dto.Text,
		dto.DueDate,
	)
	if err != nil {
		return nil, err
	}
	if dto.Done {
		item.Complete()
	}
	return item, nil
}

func mapMetadataToDomain(dto MetadataDTO) domain.Metadata {
	return domain.NewMetadata(dto.Tags, dto.Links, dto.ExternalRefs)
}

func mapTranscriptToDomain(meetingID domain.MeetingID, dto TranscriptResponse) *domain.Transcript {
	utterances := make([]domain.Utterance, len(dto.Utterances))
	for i, u := range dto.Utterances {
		utterances[i] = domain.NewUtterance(u.Speaker, u.Text, u.Timestamp, u.Confidence)
	}
	t := domain.NewTranscript(meetingID, utterances)
	return &t
}
