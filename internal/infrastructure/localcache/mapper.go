package localcache

import (
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// Mapper translates between Granola local cache DTOs and domain types.
// This is the anti-corruption layer — ensuring that local cache structure
// never leaks into our domain model.

const cacheTimestampLayout = time.RFC3339

func mapDocumentToDomain(doc CacheDocument, meta *CacheMeetingMeta) (*domain.Meeting, error) {
	createdAt, err := time.Parse(cacheTimestampLayout, doc.CreatedAt)
	if err != nil {
		createdAt = time.Now().UTC()
	}

	source := domain.SourceOther
	var participants []domain.Participant

	if meta != nil {
		source = mapConferenceToSource(meta.Conference)

		if meta.Organizer != nil && (meta.Organizer.Name != "" || meta.Organizer.Email != "") {
			participants = append(participants, domain.NewParticipant(
				meta.Organizer.Name, meta.Organizer.Email, domain.RoleHost,
			))
		}

		for _, a := range meta.Attendees {
			// Skip organizer to avoid duplicate
			if meta.Organizer != nil && a.Email == meta.Organizer.Email && a.Name == meta.Organizer.Name {
				continue
			}
			participants = append(participants, domain.NewParticipant(a.Name, a.Email, domain.RoleAttendee))
		}
	}

	mtg, err := domain.New(
		domain.MeetingID(doc.ID),
		doc.Title,
		createdAt,
		source,
		participants,
	)
	if err != nil {
		return nil, err
	}

	// Clear creation event — this is reconstitution, not creation
	mtg.ClearDomainEvents()

	// Map ProseMirror notes to summary
	if len(doc.NotesProsemirror) > 0 {
		summaryContent := prosemirrorToPlainText(doc.NotesProsemirror)
		if summaryContent != "" {
			mtg.AttachSummary(domain.NewSummary(
				domain.MeetingID(doc.ID),
				summaryContent,
				domain.SummaryAuto,
			))
			mtg.ClearDomainEvents()
		}
	}

	return mtg, nil
}

func mapTranscriptToDomain(meetingID string, transcript CacheTranscript) *domain.Transcript {
	if len(transcript.Segments) == 0 {
		return nil
	}

	utterances := make([]domain.Utterance, 0, len(transcript.Segments))
	for _, seg := range transcript.Segments {
		ts, err := time.Parse(cacheTimestampLayout, seg.Timestamp)
		if err != nil {
			ts = time.Time{}
		}
		utterances = append(utterances, domain.NewUtterance(seg.Speaker, seg.Text, ts, 0))
	}

	t := domain.NewTranscript(domain.MeetingID(meetingID), utterances)
	return &t
}

func mapConferenceToSource(conf *CacheConference) domain.Source {
	if conf == nil {
		return domain.SourceOther
	}
	switch conf.Type {
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
