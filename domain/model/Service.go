package model

type Service struct {
	Id string
}

func (s Service) AsResource() Resource {
	return Resource{Type: "service", ID: s.Id}
}
