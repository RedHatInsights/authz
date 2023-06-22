package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/testenv"
	"fmt"
	"os"
	"testing"
	"time"

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
	evts, err := repo.Connect()
	assert.NoError(t, err)
	//When
	localBrokerContainer.SendSubjectUpdated(sent)
	//Then
	received := <-evts.SubjectChanges

	assert.Equal(t, sent, received)
	assertNoErrors(t, evts.Errors)
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
