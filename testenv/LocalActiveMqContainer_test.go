package testenv

import (
	"authz/domain/contracts"
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/golang/glog"
	"github.com/stretchr/testify/assert"
)

var localContainer *LocalActiveMqContainer

func TestPublishSubjectAddedEvent(t *testing.T) {
	t.SkipNow() //Skipped pending a local test mechanism
	//Given
	expectedXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
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
					   <Identifier system="WEB" entity-name="User" qualifier="id">52915708</Identifier>
					   <Reference system="WEB" entity-name="Customer" qualifier="id">6340056</Reference>
					   <Reference system="EBS" entity-name="Account" qualifier="number">1460290</Reference>
				   </Identifiers>
				   <Status primary="true">
					   <State>Active</State>
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
	</CanonicalMessage>`

	receiver, err := localContainer.CreateReciever(umbUserEventsTopic)
	assert.NoError(t, err)
	//When
	err = localContainer.SendSubjectAdded(contracts.SubjectAddOrUpdateEvent{
		SubjectID: "52915708",
		OrgID:     "6340056",
		Active:    true,
	})
	assert.NoError(t, err)

	msg, err := receiver.Receive(context.TODO(), nil)
	assert.NoError(t, err)
	//Then
	data := string(msg.GetData())
	assert.Equal(t, expectedXML, data)

	err = receiver.Close(context.TODO())
	assert.NoError(t, err)
}

func TestPublishSubjectModifiedEvent(t *testing.T) {
	t.SkipNow() //Skipped pending a local test mechanism
	//Given
	expectedXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
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
					   <Identifier system="WEB" entity-name="User" qualifier="id">52915708</Identifier>
					   <Reference system="WEB" entity-name="Customer" qualifier="id">6340056</Reference>
					   <Reference system="EBS" entity-name="Account" qualifier="number">1460290</Reference>
				   </Identifiers>
				   <Status primary="true">
					   <State>Inactive</State>
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
	</CanonicalMessage>`

	receiver, err := localContainer.CreateReciever(umbUserEventsTopic)
	assert.NoError(t, err)
	//When
	err = localContainer.SendSubjectUpdated(contracts.SubjectAddOrUpdateEvent{
		SubjectID: "52915708",
		OrgID:     "6340056",
		Active:    false,
	})
	assert.NoError(t, err)
	msg, err := receiver.Receive(context.TODO(), nil)
	assert.NoError(t, err)
	//Then
	data := string(msg.GetData())
	assert.Equal(t, expectedXML, data)

	err = receiver.Close(context.TODO())
	assert.NoError(t, err)
}

func SkipTestMain(m *testing.M) { //Skipped pending a local test mechanism
	var err error
	factory := NewLocalActiveMqContainerFactory()
	localContainer, err = factory.CreateContainer()
	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		localContainer.Close()
		os.Exit(1)
	}

	m.Run()

	localContainer.Close()
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
		sender, err := session.NewSender(ctx, umbUserEventsTopic, nil)
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
