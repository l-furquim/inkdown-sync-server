package service

import (
	"errors"
	"testing"

	"inkdown-sync-server/internal/domain"
)

type mockNoteRepo struct {
	notes map[string]*domain.Note
}

func newMockNoteRepo() *mockNoteRepo {
	return &mockNoteRepo{
		notes: make(map[string]*domain.Note),
	}
}

func (m *mockNoteRepo) Create(note *domain.Note) error {
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepo) FindByID(id string) (*domain.Note, error) {
	if n, exists := m.notes[id]; exists {
		return n, nil
	}
	return nil, errors.New("note not found")
}

func (m *mockNoteRepo) List(userID string) ([]*domain.Note, error) {
	var notes []*domain.Note
	for _, n := range m.notes {
		if n.UserID == userID && !n.IsDeleted {
			notes = append(notes, n)
		}
	}
	return notes, nil
}

func (m *mockNoteRepo) Update(note *domain.Note) error {
	if _, exists := m.notes[note.ID]; exists {
		m.notes[note.ID] = note
		return nil
	}
	return errors.New("note not found")
}

func (m *mockNoteRepo) Delete(id string) error {
	if n, exists := m.notes[id]; exists {
		n.IsDeleted = true
		return nil
	}
	return errors.New("note not found")
}

func (m *mockNoteRepo) ListByWorkspace(workspaceID string) ([]*domain.Note, error) {
	var notes []*domain.Note
	for _, n := range m.notes {
		if n.WorkspaceID == workspaceID && !n.IsDeleted {
			notes = append(notes, n)
		}
	}
	return notes, nil
}

type mockVersionRepo struct{}

func (m *mockVersionRepo) SaveVersion(note *domain.Note) error { return nil }
func (m *mockVersionRepo) GetVersions(noteID string, limit int) ([]*domain.NoteVersion, error) {
	return nil, nil
}
func (m *mockVersionRepo) GetVersion(noteID string, version int64) (*domain.NoteVersion, error) {
	return nil, nil
}
func (m *mockVersionRepo) DeleteOldVersions(noteID string, keepLast int) error { return nil }

func TestNoteService_Create(t *testing.T) {
	repo := newMockNoteRepo()
	versionRepo := &mockVersionRepo{}
	service := NewNoteService(repo, versionRepo, nil, nil)

	req := &domain.CreateNoteRequest{
		Type:             domain.NoteTypeFile,
		EncryptedTitle:   "enc-title",
		EncryptedContent: "enc-content",
		EncryptionAlgo:   "AES-256-GCM",
		Nonce:            "nonce",
		DeviceID:         "device1",
	}

	note, err := service.Create("user1", req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if note.ID == "" {
		t.Error("expected note ID to be generated")
	}
	if note.Version != 1 {
		t.Errorf("expected version 1, got %d", note.Version)
	}
}

func TestNoteService_List(t *testing.T) {
	repo := newMockNoteRepo()
	versionRepo := &mockVersionRepo{}
	service := NewNoteService(repo, versionRepo, nil, nil)

	service.Create("user1", &domain.CreateNoteRequest{Type: domain.NoteTypeFile, EncryptedTitle: "n1", EncryptionAlgo: "algo", Nonce: "n", DeviceID: "d1"})
	service.Create("user1", &domain.CreateNoteRequest{Type: domain.NoteTypeFile, EncryptedTitle: "n2", EncryptionAlgo: "algo", Nonce: "n", DeviceID: "d1"})
	service.Create("user2", &domain.CreateNoteRequest{Type: domain.NoteTypeFile, EncryptedTitle: "n3", EncryptionAlgo: "algo", Nonce: "n", DeviceID: "d2"})

	list, err := service.List("user1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(list) != 2 {
		t.Errorf("expected 2 notes, got %d", len(list))
	}
}

func TestNoteService_Update(t *testing.T) {
	repo := newMockNoteRepo()
	versionRepo := &mockVersionRepo{}
	service := NewNoteService(repo, versionRepo, nil, nil)

	note, _ := service.Create("user1", &domain.CreateNoteRequest{Type: domain.NoteTypeFile, EncryptedTitle: "old", EncryptionAlgo: "algo", Nonce: "n", DeviceID: "d1"})

	newTitle := "new-enc-title"
	req := &domain.UpdateNoteRequest{
		EncryptedTitle: &newTitle,
		DeviceID:       "d1",
	}

	updated, err := service.Update("user1", note.ID, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.EncryptedTitle != newTitle {
		t.Errorf("expected title %s, got %s", newTitle, updated.EncryptedTitle)
	}
	if updated.Version != 2 {
		t.Errorf("expected version 2, got %d", updated.Version)
	}

	_, err = service.Update("user2", note.ID, req)
	if err == nil {
		t.Error("expected unauthorized error")
	}
}

func TestNoteService_Delete(t *testing.T) {
	repo := newMockNoteRepo()
	versionRepo := &mockVersionRepo{}
	service := NewNoteService(repo, versionRepo, nil, nil)

	note, _ := service.Create("user1", &domain.CreateNoteRequest{Type: domain.NoteTypeFile, EncryptedTitle: "del", EncryptionAlgo: "algo", Nonce: "n", DeviceID: "d1"})

	err := service.Delete("user1", note.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	n, _ := repo.FindByID(note.ID)
	if !n.IsDeleted {
		t.Error("expected note to be marked deleted")
	}
}
