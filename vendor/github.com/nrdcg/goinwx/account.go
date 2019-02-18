package goinwx

const (
	methodAccountLogin  = "account.login"
	methodAccountLogout = "account.logout"
	methodAccountLock   = "account.lock"
	methodAccountUnlock = "account.unlock"
)

// AccountService API access to Account.
type AccountService service

// Login Account login.
func (s *AccountService) Login() error {
	req := s.client.NewRequest(methodAccountLogin, map[string]interface{}{
		"user": s.client.username,
		"pass": s.client.password,
	})

	_, err := s.client.Do(*req)
	return err
}

// Logout Account logout.
func (s *AccountService) Logout() error {
	req := s.client.NewRequest(methodAccountLogout, nil)

	_, err := s.client.Do(*req)
	return err
}

// Lock Account lock.
func (s *AccountService) Lock() error {
	req := s.client.NewRequest(methodAccountLock, nil)

	_, err := s.client.Do(*req)
	return err
}

// Unlock Account unlock.
func (s *AccountService) Unlock(tan string) error {
	req := s.client.NewRequest(methodAccountUnlock, map[string]interface{}{
		"tan": tan,
	})

	_, err := s.client.Do(*req)
	return err
}
