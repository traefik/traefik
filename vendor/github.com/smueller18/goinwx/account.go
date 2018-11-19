package goinwx

const (
	methodAccountLogin  = "account.login"
	methodAccountLogout = "account.logout"
	methodAccountLock   = "account.lock"
	methodAccountUnlock = "account.unlock"
)

type AccountService interface {
	Login() error
	Logout() error
	Lock() error
	Unlock(tan string) error
}

type AccountServiceOp struct {
	client *Client
}

var _ AccountService = &AccountServiceOp{}

func (s *AccountServiceOp) Login() error {
	req := s.client.NewRequest(methodAccountLogin, map[string]interface{}{
		"user": s.client.Username,
		"pass": s.client.Password,
	})

	_, err := s.client.Do(*req)
	return err
}

func (s *AccountServiceOp) Logout() error {
	req := s.client.NewRequest(methodAccountLogout, nil)

	_, err := s.client.Do(*req)
	return err
}

func (s *AccountServiceOp) Lock() error {
	req := s.client.NewRequest(methodAccountLock, nil)

	_, err := s.client.Do(*req)
	return err
}

func (s *AccountServiceOp) Unlock(tan string) error {
	req := s.client.NewRequest(methodAccountUnlock, map[string]interface{}{
		"tan": tan,
	})

	_, err := s.client.Do(*req)
	return err
}
