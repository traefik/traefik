package project

import (
	"sync"

	"github.com/docker/libcompose/project/events"
	log "github.com/sirupsen/logrus"
)

type serviceWrapper struct {
	name    string
	service Service
	done    sync.WaitGroup
	state   ServiceState
	err     error
	project *Project
	noWait  bool
	ignored map[string]bool
}

func newServiceWrapper(name string, p *Project) (*serviceWrapper, error) {
	wrapper := &serviceWrapper{
		name:    name,
		state:   StateUnknown,
		project: p,
		ignored: map[string]bool{},
	}

	return wrapper, wrapper.Reset()
}

func (s *serviceWrapper) IgnoreDep(name string) {
	s.ignored[name] = true
}

func (s *serviceWrapper) Reset() error {
	if s.state != StateExecuted {
		service, err := s.project.CreateService(s.name)
		if err != nil {
			log.Errorf("Failed to create service for %s : %v", s.name, err)
			return err
		}

		s.service = service
	}

	if s.err == ErrRestart {
		s.err = nil
	}
	s.done.Add(1)

	return nil
}

func (s *serviceWrapper) Ignore() {
	defer s.done.Done()

	s.state = StateExecuted
	s.project.Notify(events.ServiceUpIgnored, s.service.Name(), nil)
}

func (s *serviceWrapper) waitForDeps(wrappers map[string]*serviceWrapper) bool {
	if s.noWait {
		return true
	}

	for _, dep := range s.service.DependentServices() {
		if s.ignored[dep.Target] {
			continue
		}

		if wrapper, ok := wrappers[dep.Target]; ok {
			if wrapper.Wait() == ErrRestart {
				s.project.Notify(events.ProjectReload, wrapper.service.Name(), nil)
				s.err = ErrRestart
				return false
			}
		} else {
			log.Errorf("Failed to find %s", dep.Target)
		}
	}

	return true
}

func (s *serviceWrapper) Do(wrappers map[string]*serviceWrapper, start, done events.EventType, action func(service Service) error) {
	defer s.done.Done()

	if s.state == StateExecuted {
		return
	}

	if wrappers != nil && !s.waitForDeps(wrappers) {
		return
	}

	s.state = StateExecuted

	s.project.Notify(start, s.service.Name(), nil)

	s.err = action(s.service)
	if s.err == ErrRestart {
		s.project.Notify(done, s.service.Name(), nil)
		s.project.Notify(events.ProjectReloadTrigger, s.service.Name(), nil)
	} else if s.err != nil {
		log.Errorf("Failed %s %s : %v", start, s.name, s.err)
	} else {
		s.project.Notify(done, s.service.Name(), nil)
	}
}

func (s *serviceWrapper) Wait() error {
	s.done.Wait()
	return s.err
}
