package application

import (
	"authz/infrastructure/repository/authzed"
	"fmt"
	"os"
	"testing"
)

var spicedbContainer *authzed.LocalSpiceDbContainer

func TestMain(m *testing.M) {
	factory := authzed.NewLocalSpiceDbContainerFactory()
	var err error
	spicedbContainer, err = factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		os.Exit(-1)
	}

	result := m.Run()

	spicedbContainer.Close()
	os.Exit(result)
}
