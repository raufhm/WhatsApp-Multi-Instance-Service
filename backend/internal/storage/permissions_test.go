package storage

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestLogPermissionCheck(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	operatorID := uuid.New()
	resourceID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO operator_permission_checks")).
		WithArgs(operatorID, "operators:create", "operators", resourceID, false, "denied", "127.0.0.1", "test-agent").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.LogPermissionCheck(context.Background(), operatorID, "operators:create", "operators", resourceID, false, "denied", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("LogPermissionCheck: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestLogPermissionCheckNulls(t *testing.T) {
	store, mock, closeDB := newProjectionStore(t)
	defer closeDB()

	operatorID := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO operator_permission_checks")).
		WithArgs(operatorID, "operators:view", nil, nil, true, "allowed", nil, nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := store.LogPermissionCheck(context.Background(), operatorID, "operators:view", "", uuid.Nil, true, "allowed", "", "")
	if err != nil {
		t.Fatalf("LogPermissionCheck: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
