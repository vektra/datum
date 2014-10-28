package datum

import "github.com/stretchr/testify/mock"

type MockBlobStore struct {
	mock.Mock
}

func (m *MockBlobStore) Set(key string, space string, val []byte) error {
	ret := m.Called(key, space, val)

	r0 := ret.Error(0)

	return r0
}
func (m *MockBlobStore) Get(key string, space string) ([]byte, error) {
	ret := m.Called(key, space)

	r0 := ret.Get(0).([]byte)
	r1 := ret.Error(1)

	return r0, r1
}
