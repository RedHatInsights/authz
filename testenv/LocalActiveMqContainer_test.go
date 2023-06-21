package testenv

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestContainerInitialization(t *testing.T) {
	factory := NewLocalActiveMqContainerFactory()
	start := time.Now()
	container, err := factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		container.Close()
		os.Exit(1)
	}
	elapsed := time.Since(start).Seconds()
	fmt.Printf("CONNECTION INITIALIZED AFTER %f Seconds\n", elapsed)
	container.Close()
}
