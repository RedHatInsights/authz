//go:build !release

package testenv

import (
	"authz/domain/contracts"
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/golang/glog"

	"github.com/ory/dockertest"
)

// LocalActiveMqContainerFactory is only used for test setup and not included in builds with the release tag
type LocalActiveMqContainerFactory struct {
}

// LocalActiveMqContainer struct that holds pointers to the container, dockertest pool and exposes the port
type LocalActiveMqContainer struct {
	mgmtPort  string
	amqpPort  string
	container *dockertest.Resource
	sender    *amqp.Sender
	conn      *amqp.Conn
	pool      *dockertest.Pool
}

// NewLocalActiveMqContainerFactory constructor for the factory
func NewLocalActiveMqContainerFactory() *LocalActiveMqContainerFactory {
	return &LocalActiveMqContainerFactory{}
}

// CreateContainer creates a new SpiceDbContainer using dockertest
func (l *LocalActiveMqContainerFactory) CreateContainer() (*LocalActiveMqContainer, error) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	pool.MaxWait = 3 * time.Minute

	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "quay.io/artemiscloud/activemq-artemis-broker",
		Tag:        "latest",
		Env: []string{
			"AMQ_USER=admin",
			"AMQ_PASSWORD=admin",
		},
		Mounts: []string{
			path.Join(basepath, "../testdata/activemq/bootstrap.xml") + ":/var/lib/artemis/etc/bootstrap.xml",
			path.Join(basepath, "../testdata/activemq/broker.xml") + ":/var/lib/artemis/etc/broker.xml",
			path.Join(basepath, "../testdata/activemq/login.config") + ":/var/lib/artemis/etc/login.config",
			path.Join(basepath, "../testdata/activemq/roles.properties") + ":/var/lib/artemis/etc/artemis-roles.properties",
			path.Join(basepath, "../testdata/activemq/users.properties") + ":/var/lib/artemis/etc/artemis-users.properties",
		},
		ExposedPorts: []string{"61616/tcp", "5672/tcp", "8161/tcp"},
	})

	if err != nil {
		return nil, fmt.Errorf("could not start activeMQ resource: %w", err)
	}

	mgmtPort := resource.GetPort("8161/tcp")
	amqpPort := resource.GetPort("5672/tcp")

	// Give the service time to boot.
	cErr := pool.Retry(func() error {
		log.Print("Attempting to connect to activeMQ...")

		result, err := http.Get(fmt.Sprintf("http://localhost:%s/console", mgmtPort))
		_ = result
		if err != nil {
			return fmt.Errorf("error connecting to acrtiveMQ: %v", err.Error())
		}

		return err
	})

	if cErr != nil {
		return nil, cErr
	}

	conn, sender, err := createSender("amqp://localhost:" + amqpPort)
	if err != nil {
		if conn != nil {
			err = conn.Close()
			if err != nil {
				glog.Errorf("Failed to close connection: %v", err)
			}
		}
	}

	return &LocalActiveMqContainer{
		mgmtPort:  mgmtPort,
		amqpPort:  amqpPort,
		sender:    sender,
		conn:      conn,
		container: resource,
		pool:      pool,
	}, nil
}

func createSender(url string) (conn *amqp.Conn, sender *amqp.Sender, err error) {
	ctx := context.TODO()

	// create connection
	conn, err = amqp.Dial(ctx, url, &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("writer", "password2"),
	})
	if err != nil {
		return
	}

	// open a session
	session, err := conn.NewSession(ctx, nil)
	if err != nil {
		return
	}

	// create a sender
	sender, err = session.NewSender(ctx, "testTopic", nil)

	return
}

// SendSubjectAdded sends a SubjectAddOrUpdateEvent representing a new subject to the local container
func (l *LocalActiveMqContainer) SendSubjectAdded(evt contracts.SubjectAddOrUpdateEvent) error {
	msg := amqp.NewMessage(createSubjectCreatedEventData(evt.SubjectID, evt.OrgID, evt.Active))
	return l.sender.Send(context.TODO(), msg, nil)
}

// SendSubjectUpdated sends a SubjectAddOrUpdateEvent representing a modified subject to the local container
func (l *LocalActiveMqContainer) SendSubjectUpdated(evt contracts.SubjectAddOrUpdateEvent) error {
	msg := amqp.NewMessage(createSubjectUpdatedEventData(evt.SubjectID, evt.OrgID, evt.Active))
	return l.sender.Send(context.TODO(), msg, nil)
}

func createSubjectUpdatedEventData(subjectID string, orgID string, active bool) []byte {
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<CanonicalMessage xmlns="http://esb.redhat.com/Canonical/6">
	   <Header>
		   <System>WEB</System>
		   <Operation>update</Operation>
		   <Type>User</Type>
		   <InstanceId>5e86560b88975300017006aa</InstanceId>
		   <Timestamp>2020-04-02T17:15:59.906</Timestamp>
	   </Header>
	   <Payload>
		   <Sync>
			   <User>
				   <CreatedDate>2020-04-02T17:10:46.856</CreatedDate>
				   <LastUpdatedDate>2020-04-02T17:15:54.767</LastUpdatedDate>
				   <Identifiers>
					   <Identifier system="WEB" entity-name="User" qualifier="id">%s</Identifier>
					   <Reference system="WEB" entity-name="Customer" qualifier="id">%s</Reference>
					   <Reference system="EBS" entity-name="Account" qualifier="number">1460290</Reference>
				   </Identifiers>
				   <Status primary="true">
					   <State>%s</State>
				   </Status>
				   <Person>
					   <FirstName>firstName</FirstName>
					   <LastName>lastName</LastName>
					   <Title>jobTitle</Title>
					   <Credentials>
						   <Login>test-principal-1234546</Login>
					   </Credentials>
				   </Person>
				   <Company>
					   <Name>Red Hat Inc.</Name>
				   </Company>
				   <Address>
					   <Identifiers>
						   <AuthoringOperatingUnit>
							   <Number>103</Number>
						   </AuthoringOperatingUnit>
						   <Identifier system="WEB" entity-name="Address" entity-type="Customer Site" qualifier="id">28787516_SITE</Identifier>
					   </Identifiers>
					   <Status primary="true">
						   <State>Active</State>
					   </Status>
					   <Line number="1">100 East Davie Street</Line>
					   <City>RALEIGH</City>
					   <Subdivision type="County">WAKE</Subdivision>
					   <State>NC</State>
					   <CountryISO2Code>US</CountryISO2Code>
					   <PostalCode>27601</PostalCode>
				   </Address>
				   <Phone type="Gen" primary="true">
					   <Identifiers>
						   <Identifier system="WEB" entity-name="Phone" qualifier="id">52915708_IPHONE</Identifier>
					   </Identifiers>
					   <Number>1-919-754-4950</Number>
					   <RawNumber>1-919-754-4950</RawNumber>
				   </Phone>
				   <Email primary="true">
					   <Identifiers>
						   <Identifier system="WEB" entity-name="Email" qualifier="id">52915708_IEMAIL</Identifier>
					   </Identifiers>
					   <EmailAddress>test@redhat.com</EmailAddress>
				   </Email>
				   <UserMembership>
					   <Name>redhat:employees</Name>
				   </UserMembership>
				   <UserPrivilege>
					   <Label>portal_system_management</Label>
					   <Description>Customer Portal: System Management</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_download</Label>
					   <Description>Customer Portal: Download Software and Updates</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_manage_subscriptions</Label>
					   <Description>Customer Portal: Manage Subscriptions</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_manage_cases</Label>
					   <Description>Customer Portal: Manage Support Cases</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
			   </User>
		   </Sync>
	   </Payload>
	</CanonicalMessage>`, subjectID, orgID, convertActiveToString(active)))
}

func createSubjectCreatedEventData(subjectID string, orgID string, active bool) []byte {
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
	<CanonicalMessage xmlns="http://esb.redhat.com/Canonical/6">
	   <Header>
		   <System>WEB</System>
		   <Operation>insert</Operation>
		   <Type>User</Type>
		   <InstanceId>5e8654da889753000170061a</InstanceId>
		   <Timestamp>2020-04-02T17:11:00.936</Timestamp>
	   </Header>
	   <Payload>
		   <Sync>
			   <User>
				   <CreatedDate>2020-04-02T17:10:46.856</CreatedDate>
				   <LastUpdatedDate>2020-04-02T17:10:47.321</LastUpdatedDate>
				   <Identifiers>
					   <Identifier system="WEB" entity-name="User" qualifier="id">%s</Identifier>
					   <Reference system="WEB" entity-name="Customer" qualifier="id">%s</Reference>
					   <Reference system="EBS" entity-name="Account" qualifier="number">1460290</Reference>
				   </Identifiers>
				   <Status primary="true">
					   <State>%s</State>
				   </Status>
				   <Person>
					   <FirstName>firstName</FirstName>
					   <LastName>lastName</LastName>
					   <Title>jobTitle</Title>
					   <Credentials>
						   <Login>test-principal-1234546</Login>
					   </Credentials>
				   </Person>
				   <Company>
					   <Name>Red Hat Inc.</Name>
				   </Company>
				   <Address>
					   <Identifiers>
						   <AuthoringOperatingUnit>
							   <Number>103</Number>
						   </AuthoringOperatingUnit>
						   <Identifier system="WEB" entity-name="Address" entity-type="Customer Site" qualifier="id">28787516_SITE</Identifier>
					   </Identifiers>
					   <Status primary="true">
						   <State>Active</State>
					   </Status>
					   <Line number="1">100 East Davie Street</Line>
					   <City>RALEIGH</City>
					   <Subdivision type="County">WAKE</Subdivision>
					   <State>NC</State>
					   <CountryISO2Code>US</CountryISO2Code>
					   <PostalCode>27601</PostalCode>
				   </Address>
				   <Phone type="Gen" primary="true">
					   <Identifiers>
						   <Identifier system="WEB" entity-name="Phone" qualifier="id">52915708_IPHONE</Identifier>
					   </Identifiers>
					   <Number>+1 919-754-4950</Number>
					   <RawNumber>+1 919-754-4950</RawNumber>
				   </Phone>
				   <Email primary="true">
					   <Identifiers>
						   <Identifier system="WEB" entity-name="Email" qualifier="id">52915708_IEMAIL</Identifier>
					   </Identifiers>
					   <EmailAddress>test@redhat.com</EmailAddress>
				   </Email>
				   <UserMembership>
					   <Name>redhat:employees</Name>
				   </UserMembership>
				   <UserPrivilege>
					   <Label>portal_system_management</Label>
					   <Description>Customer Portal: System Management</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_download</Label>
					   <Description>Customer Portal: Download Software and Updates</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_manage_subscriptions</Label>
					   <Description>Customer Portal: Manage Subscriptions</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
				   <UserPrivilege>
					   <Label>portal_manage_cases</Label>
					   <Description>Customer Portal: Manage Support Cases</Description>
					   <Privileged>Y</Privileged>
				   </UserPrivilege>
			   </User>
		   </Sync>
	   </Payload>
	</CanonicalMessage>`, subjectID, orgID, convertActiveToString(active)))
}

func convertActiveToString(active bool) string {
	if active {
		return "Active"
	}

	return "Inactive"
}

// AmqpPort returns the Port the container is listening
func (l *LocalActiveMqContainer) AmqpPort() string {
	return l.amqpPort
}

// Close purges the container
func (l *LocalActiveMqContainer) Close() {
	if l.conn != nil {
		err := l.conn.Close()
		if err != nil {
			glog.Errorf("Error disconnecting from container: %s", err)
		}
	}
	err := l.pool.Purge(l.container)
	if err != nil {
		glog.Error("Could not purge activeMQ Container from test. Please delete manually.")
	}
}
