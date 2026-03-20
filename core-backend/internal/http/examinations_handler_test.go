package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dimplom/internal/examinations"
	"github.com/go-chi/chi/v5"
)

func TestFinishIsIdempotent(t *testing.T) {
	finishedAt := time.Unix(100, 0).UTC()
	handler := ExaminationsHandler{
		service: examinations.NewService(&examServiceRepoStub{
			finishResult: examinations.Examination{
				ID:           12,
				SpecialistID: 7,
				Status:       examinations.StatusReadyForProcessing,
				FinishedAt:   &finishedAt,
			},
		}),
	}

	req := httptest.NewRequest(nethttp.MethodPost, "/examinations/12/finish", nil)
	req = withURLParam(req, "id", "12")
	rec := httptest.NewRecorder()

	handler.Finish(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}
	var exam examinations.Examination
	if err := json.Unmarshal(rec.Body.Bytes(), &exam); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if exam.Status != examinations.StatusReadyForProcessing {
		t.Fatalf("expected ready_for_processing, got %s", exam.Status)
	}
}

func TestFinishRequiresAllAnswers(t *testing.T) {
	handler := ExaminationsHandler{
		service: examinations.NewService(&examServiceRepoStub{
			finishErr: examinations.ErrAnswersIncomplete,
		}),
	}

	req := httptest.NewRequest(nethttp.MethodPost, "/examinations/12/finish", nil)
	req = withURLParam(req, "id", "12")
	rec := httptest.NewRecorder()

	handler.Finish(rec, req)

	if rec.Code != nethttp.StatusConflict {
		t.Fatalf("expected status %d, got %d", nethttp.StatusConflict, rec.Code)
	}
}

func TestListBySpecialistReturnsCurrentStatuses(t *testing.T) {
	startedAt := time.Unix(100, 0).UTC()
	finishedAt := time.Unix(200, 0).UTC()
	handler := ExaminationsHandler{
		service: examinations.NewService(&examServiceRepoStub{
			listBySpecialistResult: []examinations.Examination{
				{ID: 1, SpecialistID: 7, Status: examinations.StatusCreated},
				{ID: 2, SpecialistID: 7, Status: examinations.StatusCollectingAnswers, StartedAt: &startedAt},
				{ID: 3, SpecialistID: 7, Status: examinations.StatusReadyForProcessing, StartedAt: &startedAt, FinishedAt: &finishedAt},
			},
		}),
	}

	req := httptest.NewRequest(nethttp.MethodGet, "/specialists/7/examinations", nil)
	req = withURLParam(req, "id", "7")
	rec := httptest.NewRecorder()

	handler.ListBySpecialist(rec, req)

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected status %d, got %d", nethttp.StatusOK, rec.Code)
	}

	var payload examinationsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(payload.Items))
	}
	if payload.Items[0].Status != examinations.StatusCreated ||
		payload.Items[1].Status != examinations.StatusCollectingAnswers ||
		payload.Items[2].Status != examinations.StatusReadyForProcessing {
		t.Fatal("expected authoritative status list in response")
	}
}

type examServiceRepoStub struct {
	createResult           examinations.Examination
	createErr              error
	getByIDResult          examinations.Examination
	getByIDErr             error
	listBySpecialistResult []examinations.Examination
	listBySpecialistErr    error
	updateResult           examinations.Examination
	updateErr              error
	finishResult           examinations.Examination
	finishErr              error
}

func (s *examServiceRepoStub) Create(context.Context, examinations.CreateInput) (examinations.Examination, error) {
	return s.createResult, s.createErr
}

func (s *examServiceRepoStub) List(context.Context) ([]examinations.Examination, error) {
	return nil, nil
}

func (s *examServiceRepoStub) GetByID(context.Context, int64) (examinations.Examination, error) {
	return s.getByIDResult, s.getByIDErr
}

func (s *examServiceRepoStub) ListBySpecialistID(context.Context, int64) ([]examinations.Examination, error) {
	return s.listBySpecialistResult, s.listBySpecialistErr
}

func (s *examServiceRepoStub) UpdateStatus(context.Context, int64, string) (examinations.Examination, error) {
	return s.updateResult, s.updateErr
}

func (s *examServiceRepoStub) Finish(context.Context, int64) (examinations.Examination, error) {
	return s.finishResult, s.finishErr
}

func withURLParam(req *nethttp.Request, key, value string) *nethttp.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}
