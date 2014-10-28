package datum

import "github.com/stretchr/testify/mock"

type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) NewToken() string {
	ret := m.Called()

	r0 := ret.Get(0).(string)

	return r0
}
