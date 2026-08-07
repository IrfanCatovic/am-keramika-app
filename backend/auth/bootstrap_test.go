package auth

import (
	"testing"

	"am-keramika-backend/database"
	"am-keramika-backend/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupBootstrapDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB = db
}

func TestEnsureInitialDeveloperCreatesHashedAccount(t *testing.T) {
	setupBootstrapDB(t)
	t.Setenv("ENABLE_DEVELOPER_BOOTSTRAP", "true")
	t.Setenv("INITIAL_DEVELOPER_USERNAME", "  DevOwner  ")
	t.Setenv("INITIAL_DEVELOPER_PASSWORD", "strongpass1")

	if err := EnsureInitialDeveloper(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	var user models.User
	if err := database.DB.Where("username = ?", "devowner").First(&user).Error; err != nil {
		t.Fatalf("find developer: %v", err)
	}
	if user.Role != models.RoleDeveloper {
		t.Fatalf("role want developer got %s", user.Role)
	}
	if !user.IsActive {
		t.Fatal("expected active developer")
	}
	if user.PasswordHash == "" || user.PasswordHash == "strongpass1" {
		t.Fatal("expected bcrypt hash, not plain password")
	}
	if !CheckPassword(user.PasswordHash, "strongpass1") {
		t.Fatal("password hash does not match")
	}
	if user.FullName != "Irfan Catovic" {
		t.Fatalf("developer fullName want Irfan Catovic got %q", user.FullName)
	}
}

func TestEnsureInitialDeveloperIdempotentDoesNotResetPassword(t *testing.T) {
	setupBootstrapDB(t)
	t.Setenv("ENABLE_DEVELOPER_BOOTSTRAP", "true")
	t.Setenv("INITIAL_DEVELOPER_USERNAME", "devowner")
	t.Setenv("INITIAL_DEVELOPER_PASSWORD", "strongpass1")

	if err := EnsureInitialDeveloper(); err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}

	var first models.User
	if err := database.DB.Where("username = ?", "devowner").First(&first).Error; err != nil {
		t.Fatalf("find: %v", err)
	}

	t.Setenv("INITIAL_DEVELOPER_PASSWORD", "completely-new-password")
	if err := EnsureInitialDeveloper(); err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}

	var second models.User
	if err := database.DB.Where("username = ?", "devowner").First(&second).Error; err != nil {
		t.Fatalf("find after second: %v", err)
	}

	var count int64
	if err := database.DB.Model(&models.User{}).Where("role = ?", models.RoleDeveloper).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 developer, got %d", count)
	}
	if first.PasswordHash != second.PasswordHash {
		t.Fatal("second startup must not reset developer password")
	}
	if !CheckPassword(second.PasswordHash, "strongpass1") {
		t.Fatal("original password should still work")
	}
	if CheckPassword(second.PasswordHash, "completely-new-password") {
		t.Fatal("new env password must not replace hash")
	}
}

func TestEnsureInitialDeveloperRequiresBootstrapWhenMissing(t *testing.T) {
	setupBootstrapDB(t)
	t.Setenv("ENABLE_DEVELOPER_BOOTSTRAP", "false")
	t.Setenv("INITIAL_DEVELOPER_USERNAME", "")
	t.Setenv("INITIAL_DEVELOPER_PASSWORD", "")

	err := EnsureInitialDeveloper()
	if err == nil {
		t.Fatal("expected error when no developer and bootstrap disabled")
	}
}

func TestEnsureInitialDeveloperIgnoresExistingSoftDeleted(t *testing.T) {
	setupBootstrapDB(t)
	hash, err := HashPassword("strongpass1")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	dev := models.User{
		Username:     "olddev",
		PasswordHash: hash,
		Role:         models.RoleDeveloper,
		IsActive:     false,
	}
	if err := database.DB.Create(&dev).Error; err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := database.DB.Delete(&dev).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	t.Setenv("ENABLE_DEVELOPER_BOOTSTRAP", "true")
	t.Setenv("INITIAL_DEVELOPER_USERNAME", "newdev")
	t.Setenv("INITIAL_DEVELOPER_PASSWORD", "strongpass1")

	if err := EnsureInitialDeveloper(); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	var count int64
	if err := database.DB.Unscoped().Model(&models.User{}).Where("role = ?", models.RoleDeveloper).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected existing soft-deleted developer to block create, got count %d", count)
	}
}
