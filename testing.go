package mond

type StubLogStore struct {
	AppAccessLogs Apps
}

func (s *StubLogStore) GetApps() Apps {
	return s.AppAccessLogs
}


func (s *StubLogStore) GetApp(name string) *App {
	app := s.AppAccessLogs.Find(name)
	if app != nil {
		return app
	}
	return nil
}

func (s *StubLogStore) GetAccessLogs(name string) AccessLogs {
	app := s.AppAccessLogs.Find(name)
	if app != nil {
		return app.Logs
	}
	return AccessLogs{}
}

func (s *StubLogStore) RecordAccessLog(name string, value AccessLog) {
	app := s.AppAccessLogs.Find(name)
	if app != nil {
		app.Logs = append(app.Logs, value)
	} else {
		s.AppAccessLogs = append(s.AppAccessLogs, App{
			Name: name,
			Logs: AccessLogs{value},
		})
	}
}

func (s *StubLogStore) GetAppNames() []string {
	var apps []string
	for _, v := range s.AppAccessLogs {
		apps = append(apps, v.Name)
	}
	return apps
}

func (s *StubLogStore) GetHealth(name string) HealthCheck {
	app := s.AppAccessLogs.Find(name)
	if app != nil {
		return app.Health
	}
	return HealthCheck{}
}

func (s *StubLogStore) RecordHealth(name string, check HealthCheck) {
	app := s.AppAccessLogs.Find(name)
	if app != nil {
		app.Health = check
	} else {
		s.AppAccessLogs = append(s.AppAccessLogs, App{
			Name:   name,
			Health: check,
		})
	}
}
