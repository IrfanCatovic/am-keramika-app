package handlers

import (
	"strings"

	"am-keramika-backend/dto"
	"am-keramika-backend/models"
)

// userDisplayName prefers full name for people-facing labels; username is login-only fallback.
func userDisplayName(fullName, username string) string {
	if name := strings.TrimSpace(fullName); name != "" {
		return name
	}
	return strings.TrimSpace(username)
}

func mapUserSummary(user models.User) *dto.UserSummaryResponse {
	if user.ID == 0 {
		return nil
	}
	return &dto.UserSummaryResponse{
		ID:       user.ID,
		Username: user.Username,
		FullName: strings.TrimSpace(user.FullName),
	}
}
