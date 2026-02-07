package meeting_test

import (
	"context"
	"testing"

	app "github.com/felixgeelhaar/acai/internal/application/meeting"
)

func TestSearchTranscripts_DelegatesToRepository(t *testing.T) {
	repo := newMockRepository()
	m := mustNewMeeting(t, "m-1", "Meeting with keyword")
	repo.addMeeting(m)

	uc := app.NewSearchTranscripts(repo)
	out, err := uc.Execute(context.Background(), app.SearchTranscriptsInput{
		Query: "keyword",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.searchCalled {
		t.Error("expected SearchTranscripts to be called")
	}
	if out.Total != 1 {
		t.Errorf("got total %d, want 1", out.Total)
	}
}

func TestSearchTranscripts_EmptyQuery(t *testing.T) {
	repo := newMockRepository()
	uc := app.NewSearchTranscripts(repo)

	_, err := uc.Execute(context.Background(), app.SearchTranscriptsInput{Query: ""})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}
