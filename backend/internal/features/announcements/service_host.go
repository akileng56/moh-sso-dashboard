package announcements

import announcementmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"

// Service embeds the pinned core. Compatibility setters are composition-time
// adapters, called before the host starts serving requests or workers.
type Service struct {
	*announcementmodule.Service
	repo     Repository
	ports    ServicePorts
	settings *Settings
}

func NewServiceWithSettings(repo Repository, ports ServicePorts, settings *Settings) *Service {
	return &Service{Service: announcementmodule.NewServiceWithSettings(repo, ports, settings), repo: repo, ports: ports, settings: settings}
}

func (s *Service) recompose() {
	s.Service = announcementmodule.NewServiceWithSettings(s.repo, s.ports, s.settings)
}
