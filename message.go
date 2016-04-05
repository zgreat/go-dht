package main

import "time"

// Message represents a communication packet between two nodes on the network
type Message interface {
	// the source node Key
	GetSourceKey() Key
	// the desired Key (what we're trying to reach)
	GetDesiredKey() Key

	// the command requested
	GetCommand() string
	// the time this message was created
	GetTimestamp() time.Time

	// the random message ID associated with this message
	GetID() Key

	// any additional arguments the command needs
	GetArgs() map[string]interface{}
}

type nodeMessage struct {
	Source  Key `codec:"source" json:"source"`
	Desired Key `codec:"desired" json:"desired"`

	// the message identifier
	ID Key `codec:"id" json:"id"`

	Command   string    `codec:"command" json:"command"`
	Timestamp time.Time `codec:"timestamp" json:"timestamp"`

	Args map[string]interface{} `codec:"args" json:"args"`
}

// NewMessage generates a new message object from the passed in parameters
func NewMessage(Source, Desired Key, Command string, Args map[string]interface{}) Message {
	m := &nodeMessage{
		Source:    Source,
		Desired:   Desired,
		Command:   Command,
		Timestamp: time.Now(),
		Args:      Args,
		ID:        NewKey(),
	}

	return m
}

func (n *nodeMessage) GetSourceKey() Key {
	return n.Source
}

func (n *nodeMessage) GetDesiredKey() Key {
	return n.Desired
}

func (n *nodeMessage) GetCommand() string {
	return n.Command
}

func (n *nodeMessage) GetTimestamp() time.Time {
	return n.Timestamp
}

func (n *nodeMessage) GetArgs() map[string]interface{} {
	return n.Args
}

func (n *nodeMessage) GetID() Key {
	return n.ID
}
