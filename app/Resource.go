package app

//A Resource is a securable asset Principals may or may not have authority over
type Resource struct {
	//The Type is the kind of Resource that it is
	Type string
	//IDs must be permanent and unique within a Type
	ID string
}
