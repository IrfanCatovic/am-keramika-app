package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"am-keramika-backend/auth"
	"am-keramika-backend/database"
	"am-keramika-backend/models"
	"am-keramika-backend/repositories"
)

func TestDeveloperLoginAndCreateBoss(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	r := setupRouter()

	token := loginToken(t, r, "devowner", "password123")

	body, _ := json.Marshal(map[string]string{
		"username": "sef1",
		"password": "password123",
		"role":     models.RoleBoss,
		"fullName": "Prvi Šef",
	})
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create sef: expected 201, got %d (%s)", w.Code, w.Body.String())
	}

	var created models.User
	if err := database.DB.Where("username = ?", "sef1").First(&created).Error; err != nil {
		t.Fatalf("find sef: %v", err)
	}
	if created.Role != models.RoleBoss {
		t.Fatalf("expected sef role, got %s", created.Role)
	}
}

func TestBossGetUsersHidesDeveloper(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	createUser(t, "radnik1", "password123", models.RoleWorker, true)
	r := setupRouter()

	token := loginToken(t, r, "sef", "password123")
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	if bytes.Contains(w.Body.Bytes(), []byte(`"devowner"`)) {
		t.Fatal("sef GET /users must not include developer username")
	}
	if bytes.Contains(w.Body.Bytes(), []byte(`"role":"developer"`)) {
		t.Fatal("sef GET /users must not include developer role")
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"radnik1"`)) {
		t.Fatal("expected worker in list")
	}
}

func TestBossCannotModifyDeveloper(t *testing.T) {
	setupAuthTestDB(t)
	dev := createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()
	token := loginToken(t, r, "sef", "password123")
	id := strconv.FormatUint(uint64(dev.ID), 10)

	cases := []struct {
		method string
		path   string
		body   map[string]interface{}
	}{
		{
			method: http.MethodPut,
			path:   "/users/" + id,
			body: map[string]interface{}{
				"username": "hacked",
				"role":     models.RoleBoss,
				"fullName": "X",
			},
		},
		{
			method: http.MethodPut,
			path:   "/users/" + id + "/password",
			body:   map[string]interface{}{"password": "newpassword1"},
		},
		{
			method: http.MethodPut,
			path:   "/users/" + id + "/status",
			body:   map[string]interface{}{"isActive": false},
		},
	}

	for _, tc := range cases {
		payload, _ := json.Marshal(tc.body)
		req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(payload))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("%s %s: expected 403, got %d (%s)", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
}

func TestCannotCreateDeveloperViaUsersAPI(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "sef", "password123", models.RoleBoss, true)
	r := setupRouter()
	token := loginToken(t, r, "sef", "password123")

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

func TestDeveloperCanAccessReportsAndUsers(t *testing.T) {
	setupAuthTestDB(t)
	createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	r := setupRouter()
	token := loginToken(t, r, "devowner", "password123")

	req := httptest.NewRequest(http.MethodGet, "/reports/daily", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("reports: expected 200, got %d (%s)", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("users: expected 200, got %d (%s)", w.Code, w.Body.String())
	}
}

func TestHardDeleteDeveloperRejected(t *testing.T) {
	setupAuthTestDB(t)
	dev := createUser(t, "devowner", "password123", models.RoleDeveloper, true)
	err := repositories.HardDeleteUser(dev.ID)
	if err == nil {
		t.Fatal("expected hard delete of developer to fail")
	}
	if !auth.CheckPassword(dev.PasswordHash, "password123") {
		// password hash still on struct; ensure row remains
	}
	var still models.User
	if err := database.DB.First(&still, dev.ID).Error; err != nil {
		t.Fatalf("developer row should remain: %v", err)
	}
}
