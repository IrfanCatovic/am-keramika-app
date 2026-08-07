package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"am-keramika-backend/models"
)

func TestDeveloperAndBossCanCreateAssignableRoles(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	createUser(t, "sef_test", "password123", models.RoleBoss, true)
	r := setupRouter()

	actors := []struct {
		username string
		password string
	}{
		{username: "devowner", password: "password123"},
		{username: "sef_test", password: "password123"},
	}
	roles := []string{models.RoleBoss, models.RoleManager, models.RoleWorker}

	for _, actor := range actors {
		token := loginToken(t, r, actor.username, actor.password)
		for _, role := range roles {
			username := actor.username + "_" + role
			body, _ := json.Marshal(map[string]string{
				"username": username,
				"password": "password123",
				"role":     role,
				"fullName": "Zaposleni " + role,
			})
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusCreated {
				t.Fatalf("%s create %s: expected 201, got %d (%s)", actor.username, role, w.Code, w.Body.String())
			}
			if bytes.Contains(w.Body.Bytes(), []byte(`"passwordHash"`)) ||
				bytes.Contains(w.Body.Bytes(), []byte(`"PasswordHash"`)) {
				t.Fatal("password hash must not leak in create response")
			}
		}
	}
}

func TestManagerAndWorkerForbiddenOnUsers(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	createUser(t, "menadzer1", "password123", models.RoleManager, true)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	r := setupRouter()

	for _, username := range []string{"menadzer1", "radnik1"} {
		token := loginToken(t, r, username, "password123")
		for _, methodPath := range []struct {
			method string
			path   string
			body   map[string]string
		}{
			{method: http.MethodGet, path: "/users"},
			{
				method: http.MethodPost,
				path:   "/users",
				body: map[string]string{
					"username": "x",
					"password": "password123",
					"role":     models.RoleWorker,
					"fullName": "X Radnik",
				},
			},
		} {
			var reader *bytes.Reader
			if methodPath.body != nil {
				payload, _ := json.Marshal(methodPath.body)
				reader = bytes.NewReader(payload)
			} else {
				reader = bytes.NewReader(nil)
			}
			req := httptest.NewRequest(methodPath.method, methodPath.path, reader)
			req.Header.Set("Authorization", "Bearer "+token)
			if methodPath.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("%s %s %s: expected 403, got %d (%s)", username, methodPath.method, methodPath.path, w.Code, w.Body.String())
			}
		}
	}
}

func TestDeveloperCannotCreateDeveloperViaUsersAPI(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	r := setupRouter()
	token := loginToken(t, r, "devowner", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "anotherdev",
		"password": "password123",
		"role":     models.RoleDeveloper,
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", w.Code, w.Body.String())
	}
}
