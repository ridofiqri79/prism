package model

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseUUID(value string) (pgtype.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid uuid: %w", err)
	}

	return pgtype.UUID{Bytes: parsed, Valid: true}, nil
}

func UUIDToString(value pgtype.UUID) string {
	if !value.Valid {
		return ""
	}

	return uuid.UUID(value.Bytes).String()
}
