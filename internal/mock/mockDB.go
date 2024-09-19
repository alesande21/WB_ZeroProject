package mock

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/mock"
)

type MockDBRepository struct {
	mock.Mock
}

func (m *MockDBRepository) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(*sql.Rows), argsMock.Error(1)
}

func (m *MockDBRepository) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(*sql.Row)
}

func (m *MockDBRepository) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	argsMock := m.Called(ctx, query, args)
	return argsMock.Get(0).(sql.Result), argsMock.Error(1)
}

func (m *MockDBRepository) Ping() error {
	return m.Called().Error(0)
}
