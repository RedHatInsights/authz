package model

type Organization struct {
	Id string
}

func (o Organization) AsResource() Resource {
	return Resource{Type: "organization", ID: o.Id}
}
