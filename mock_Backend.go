package config

import "github.com/stretchr/testify/mock"

type MockBackend struct {
	mock.Mock
}

func (m *MockBackend) Set(token string, space string, key string, val interface{}) error {
	ret := m.Called(token, space, key, val)

	r0 := ret.Error(0)

	return r0
}
func (m *MockBackend) Get(token string, space string, key string) (interface{}, error) {
	ret := m.Called(token, space, key)

	r0 := ret.Get(0).(interface{})
	r1 := ret.Error(1)

	return r0, r1
}
