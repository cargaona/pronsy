package denylist

import "time"

// Service the contains the methods for the domain layer.
type Service interface {
	AddDeniedDomain(string) error
	GetDeniedDomain(string) (*Denied, error)
	GetDeniedDomains() ([]Denied, error)
}

// Repository contains the methods for the Repository/Storage layer. It will be embedded within the service struct.
type Repository interface {
	AddDeniedDomain(string) error
	GetDeniedDomain(string) (*Denied, error)
	GetDeniedDomains() ([]Denied, error)
}

type Denied struct {
	Domain string    `json:"domain"`
	Date   time.Time `json:"date,omitempty"`
}

// service implements the Service interface. Also composes the Repository interface.
type service struct {
	database Repository
}

func NewService(db Repository) Service {
	return &service{database: db}
}

func (s *service) AddDeniedDomain(domain string) error {
	err := s.database.AddDeniedDomain(domain)
	if err != nil {
		return err
	}
	return nil
}

func (s *service) GetDeniedDomain(domain string) (*Denied, error) {
	response, err := s.database.GetDeniedDomain(domain)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, nil
	}
	return response, nil
}

func (s *service) GetDeniedDomains() ([]Denied, error) {
	response, err := s.database.GetDeniedDomains()
	if err != nil {
		return nil, err
	}
	return response, nil
}
