package hierarchy

type Store interface {
}

type Service struct {
	store Store
}

func NewService(repo Store) *Service {
	return &Service{store: repo}
}
