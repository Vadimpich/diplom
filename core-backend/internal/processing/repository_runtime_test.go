package processing

import (
	"testing"
	"time"
)

func TestResolveChannelRuntimeStatusUsesPublishedOutboxAsProcessingStart(t *testing.T) {
	queuedAt := time.Unix(1_742_550_000, 0).UTC()
	publishedAt := queuedAt.Add(2 * time.Second)

	status, startedAt := resolveChannelRuntimeStatus(
		"queued",
		&queuedAt,
		nil,
		nil,
		"published",
		&publishedAt,
	)

	if status != "processing" {
		t.Fatalf("expected processing status, got %q", status)
	}
	if startedAt == nil {
		t.Fatal("expected started_at to be inferred from published_at")
	}
	if !startedAt.Equal(publishedAt) {
		t.Fatalf("expected started_at=%s, got %s", publishedAt, *startedAt)
	}
}

func TestResolveChannelRuntimeStatusPreservesExplicitStartedAt(t *testing.T) {
	queuedAt := time.Unix(1_742_550_000, 0).UTC()
	startedAt := queuedAt.Add(5 * time.Second)
	publishedAt := queuedAt.Add(2 * time.Second)

	status, resolvedStartedAt := resolveChannelRuntimeStatus(
		"processing",
		&queuedAt,
		&startedAt,
		nil,
		"published",
		&publishedAt,
	)

	if status != "processing" {
		t.Fatalf("expected processing status, got %q", status)
	}
	if resolvedStartedAt == nil {
		t.Fatal("expected explicit started_at to be preserved")
	}
	if !resolvedStartedAt.Equal(startedAt) {
		t.Fatalf("expected started_at=%s, got %s", startedAt, *resolvedStartedAt)
	}
}

