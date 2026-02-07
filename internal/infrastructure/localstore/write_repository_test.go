package localstore_test

import (
	"context"
	"testing"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
	"github.com/felixgeelhaar/acai/internal/infrastructure/localstore"
)

func setupWriteRepo(t *testing.T) *localstore.WriteRepository {
	t.Helper()
	db := openTestDB(t)
	if err := localstore.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	return localstore.NewWriteRepository(db)
}

func TestWriteRepository_SaveAndGetCompleted(t *testing.T) {
	repo := setupWriteRepo(t)
	ctx := context.Background()

	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)
	item.Complete()

	if err := repo.SaveActionItemState(ctx, item); err != nil {
		t.Fatalf("save: %v", err)
	}

	found, err := repo.GetLocalActionItemState(ctx, "ai-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if found.ID() != "ai-1" {
		t.Errorf("got id %q", found.ID())
	}
	if !found.IsCompleted() {
		t.Error("should be completed")
	}
	if found.Text() != "Write report" {
		t.Errorf("got text %q", found.Text())
	}
}

func TestWriteRepository_SaveUncompleted(t *testing.T) {
	repo := setupWriteRepo(t)
	ctx := context.Background()

	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Write report", nil)

	if err := repo.SaveActionItemState(ctx, item); err != nil {
		t.Fatalf("save: %v", err)
	}

	found, err := repo.GetLocalActionItemState(ctx, "ai-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if found.IsCompleted() {
		t.Error("should not be completed")
	}
}

func TestWriteRepository_GetNotFound(t *testing.T) {
	repo := setupWriteRepo(t)
	_, err := repo.GetLocalActionItemState(context.Background(), "nonexistent")
	if err != domain.ErrMeetingNotFound {
		t.Errorf("got error %v, want %v", err, domain.ErrMeetingNotFound)
	}
}

func TestWriteRepository_SaveOverwrite(t *testing.T) {
	repo := setupWriteRepo(t)
	ctx := context.Background()

	item, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Original", nil)
	if err := repo.SaveActionItemState(ctx, item); err != nil {
		t.Fatalf("save: %v", err)
	}

	item2, _ := domain.NewActionItem("ai-1", "m-1", "Alice", "Updated", nil)
	item2.Complete()
	if err := repo.SaveActionItemState(ctx, item2); err != nil {
		t.Fatalf("save updated: %v", err)
	}

	found, err := repo.GetLocalActionItemState(ctx, "ai-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if found.Text() != "Updated" {
		t.Errorf("got text %q, want %q", found.Text(), "Updated")
	}
	if !found.IsCompleted() {
		t.Error("should be completed after update")
	}
}
