package testenv

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

func TestContainerInitialization(t *testing.T) {
	factory := NewLocalActiveMqContainerFactory()
	start := time.Now()
	container, err := factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		container.Close()
		assert.FailNow(t, "Failed to initialize container")
	}
	elapsed := time.Since(start).Seconds()
	fmt.Printf("CONNECTION INITIALIZED AFTER %f Seconds\n", elapsed)

	CreateProducer(container)
	container.Close()

}

func CreateProducer(broker *LocalActiveMqContainer) {

	ctx := context.TODO()

	// create connection
	conn, err := amqp.Dial(ctx, "amqp://localhost:"+broker.AmqpPort(), &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("writer", "password2"),
	})
	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			glog.Errorf("Failed to close connection: %v", err)
		}
	}()

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	// send a message
	{
		// create a sender
		sender, err := session.NewSender(ctx, "testTopic", nil)
		if err != nil {
			log.Fatal("Creating sender link:", err)
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

		// send message
		err = sender.Send(ctx, amqp.NewMessage([]byte("Hello!")), nil)
		if err != nil {
			log.Fatal("Sending message:", err)
		}
		fmt.Print("WORKS!!!")
		err = sender.Close(ctx)
		if err != nil {
			log.Fatal("Closing sender:", err)
		}
		cancel()
	}
}
