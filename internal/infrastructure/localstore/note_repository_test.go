package localstore_test

import (
	"context"
	"testing"

	"github.com/felixgeelhaar/acai/internal/domain/annotation"
	"github.com/felixgeelhaar/acai/internal/infrastructure/localstore"
)

func setupNoteRepo(t *testing.T) *localstore.NoteRepository {
	t.Helper()
	db := openTestDB(t)
	if err := localstore.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return localstore.NewNoteRepository(db)
}

func TestNoteRepository_SaveAndFindByID(t *testing.T) {
	repo := setupNoteRepo(t)
	ctx := context.Background()

	note, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "observation")
	if err := repo.Save(ctx, note); err != nil {
		t.Fatalf("save: %v", err)
	}

	found, err := repo.FindByID(ctx, "n-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found.ID() != "n-1" {
		t.Errorf("got id %q", found.ID())
	}
	if found.MeetingID() != "m-1" {
		t.Errorf("got meeting id %q", found.MeetingID())
	}
	if found.Author() != "claude" {
		t.Errorf("got author %q", found.Author())
	}
	if found.Content() != "observation" {
		t.Errorf("got content %q", found.Content())
	}
}

func TestNoteRepository_FindByID_NotFound(t *testing.T) {
	repo := setupNoteRepo(t)
	_, err := repo.FindByID(context.Background(), "nonexistent")
	if err != annotation.ErrNoteNotFound {
		t.Errorf("got error %v, want %v", err, annotation.ErrNoteNotFound)
	}
}

func TestNoteRepository_ListByMeeting(t *testing.T) {
	repo := setupNoteRepo(t)
	ctx := context.Background()

	note1, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "first")
	note2, _ := annotation.NewAgentNote("n-2", "m-1", "gpt", "second")
	note3, _ := annotation.NewAgentNote("n-3", "m-2", "claude", "other meeting")

	for _, n := range []*annotation.AgentNote{note1, note2, note3} {
		if err := repo.Save(ctx, n); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	notes, err := repo.ListByMeeting(ctx, "m-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2", len(notes))
	}
	if notes[0].Content() != "first" {
		t.Errorf("got first note content %q", notes[0].Content())
	}
	if notes[1].Content() != "second" {
		t.Errorf("got second note content %q", notes[1].Content())
	}
}

func TestNoteRepository_ListByMeeting_Empty(t *testing.T) {
	repo := setupNoteRepo(t)
	notes, err := repo.ListByMeeting(context.Background(), "m-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(notes) != 0 {
		t.Errorf("got %d notes, want 0", len(notes))
	}
}

func TestNoteRepository_Delete(t *testing.T) {
	repo := setupNoteRepo(t)
	ctx := context.Background()

	note, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "to delete")
	if err := repo.Save(ctx, note); err != nil {
		t.Fatalf("save: %v", err)
	}

	if err := repo.Delete(ctx, "n-1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := repo.FindByID(ctx, "n-1")
	if err != annotation.ErrNoteNotFound {
		t.Errorf("got error %v, want %v", err, annotation.ErrNoteNotFound)
	}
}

func TestNoteRepository_Delete_NotFound(t *testing.T) {
	repo := setupNoteRepo(t)
	err := repo.Delete(context.Background(), "nonexistent")
	if err != annotation.ErrNoteNotFound {
		t.Errorf("got error %v, want %v", err, annotation.ErrNoteNotFound)
	}
}

func TestNoteRepository_SaveOverwrite(t *testing.T) {
	repo := setupNoteRepo(t)
	ctx := context.Background()

	note, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "original")
	if err := repo.Save(ctx, note); err != nil {
		t.Fatalf("save: %v", err)
	}

	updated, _ := annotation.NewAgentNote("n-1", "m-1", "claude", "updated")
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("save updated: %v", err)
	}

	found, err := repo.FindByID(ctx, "n-1")
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found.Content() != "updated" {
		t.Errorf("got content %q, want %q", found.Content(), "updated")
	}
}
