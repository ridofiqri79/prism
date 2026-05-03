package service

import (
	"fmt"
	"net/http"

	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type deletionDependency struct {
	relationType  string
	relationID    string
	relationLabel string
	relationPath  string
}

func deletionBlockedError(user *model.AuthUser, entityName string, dependencies []deletionDependency) *apperrors.AppError {
	details := make([]apperrors.FieldError, 0, len(dependencies))
	for _, dep := range dependencies {
		details = append(details, apperrors.FieldError{
			Field:   dep.relationType,
			Message: fmt.Sprintf("%s | %s | id=%s", dep.relationLabel, dep.relationPath, dep.relationID),
		})
	}

	if user == nil || user.Role != "ADMIN" {
		return &apperrors.AppError{
			Code:       "FORBIDDEN",
			Message:    fmt.Sprintf("%s memiliki relasi turunan. Hanya ADMIN yang dapat mengelola penghapusan record yang sudah dipakai; hapus relasi turunan terlebih dahulu.", entityName),
			StatusCode: http.StatusForbidden,
			Details:    details,
		}
	}

	return &apperrors.AppError{
		Code:       "CONFLICT",
		Message:    fmt.Sprintf("%s tidak bisa dihapus permanen karena masih memiliki relasi turunan. Hapus relasi turunan terlebih dahulu.", entityName),
		StatusCode: http.StatusConflict,
		Details:    details,
	}
}
