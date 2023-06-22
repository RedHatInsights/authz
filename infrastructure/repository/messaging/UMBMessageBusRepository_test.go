package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/testenv"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/stretchr/testify/assert"
)

var localBrokerContainer *testenv.LocalActiveMqContainer

func TestUMBMessageRepository_receives_new_user_events(t *testing.T) {
	//given
	sent := contracts.SubjectAddOrUpdateEvent{
		SubjectID: "new_user",
		OrgID:     "o1",
		Active:    true,
	}

	repo := createUMBRepository()
	defer repo.Disconnect()

	evts, err := repo.Connect()
	assert.NoError(t, err)
	//When
	localBrokerContainer.SendSubjectAdded(sent)
	//Then
	received := <-evts.SubjectChanges

	assert.Equal(t, sent, received)
	assertNoErrors(t, evts.Errors)
}

func TestUMBMessageRepository_receives_user_deactivation_events(t *testing.T) {
	//given
	sent := contracts.SubjectAddOrUpdateEvent{
		SubjectID: "u1",
		OrgID:     "o1",
		Active:    false,
	}

	repo := createUMBRepository()
	defer repo.Disconnect()

	evts, err := repo.Connect()
	assert.NoError(t, err)
	//When
	localBrokerContainer.SendSubjectUpdated(sent)
	//Then
	received := <-evts.SubjectChanges

	assert.Equal(t, sent, received)
	assertNoErrors(t, evts.Errors)
}

func TestUMBMessageRepository_receives_user_reactivation_events(t *testing.T) {
	//given
	sent := contracts.SubjectAddOrUpdateEvent{
		SubjectID: "u3",
		OrgID:     "o1",
		Active:    true,
	}

	repo := createUMBRepository()
	defer repo.Disconnect()

	evts, err := repo.Connect()
	assert.NoError(t, err)
	//When
	localBrokerContainer.SendSubjectUpdated(sent)
	//Then
	received := <-evts.SubjectChanges

	assert.Equal(t, sent, received)
	assertNoErrors(t, evts.Errors)
}

func TestUMBMessageRepository_disconnects_successfully(t *testing.T) {
	//Given
	repo := createUMBRepository()
	evts, err := repo.Connect()
	assert.NoError(t, err)

	//When
	repo.Disconnect()

	//Assert connection is not usable (no handy 'IsOpen' analogue available)
	_, err = repo.conn.NewSession(context.TODO(), nil)
	var expected *amqp.ConnError
	assert.ErrorAs(t, err, &expected) //Weird double-pointer- ErrorAs needs a pointer to an implementation of error, and amqp errors implement with a pointer receiver
	//Assert channels are closed
	_, open := <-evts.SubjectChanges
	assert.False(t, open)
	_, open = <-evts.Errors
	assert.False(t, open)
}

func createUMBRepository() *UMBMessageBusRepository {
	return NewUMBMessageBusRepository(serviceconfig.UMBConfig{
		URL:               "amqp://localhost:" + localBrokerContainer.AmqpPort(),
		UMBClientCertFile: "",
		UMBClientCertKey:  "",
		TopicName:         "testTopic",
	}) //TODO: fill in values
}

func assertNoErrors(t *testing.T, errors chan error) {
	select {
	case err := <-errors:
		assert.NoError(t, err)
	default:
	}
}

func TestMain(m *testing.M) {
	factory := testenv.NewLocalActiveMqContainerFactory()
	start := time.Now()
	var err error
	localBrokerContainer, err = factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		localBrokerContainer.Close()
		os.Exit(1)
	}
	elapsed := time.Since(start).Seconds()
	fmt.Printf("CONNECTION INITIALIZED AFTER %f Seconds\n", elapsed)

	result := m.Run()

	localBrokerContainer.Close()
	os.Exit(result)
}
