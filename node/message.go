package node

import (
	"time"

	"github.com/gearmover/gofs2/util"
	"github.com/ugorji/go/codec"
)

// Message represents a communication packet between two nodes on the network
type Message interface {
	// the source node Key
	GetSourceKey() util.Key
	// the desired Key (what we're trying to reach)
	GetDesiredKey() util.Key

	// the command requested
	GetCommand() string
	// the time this message was created
	GetTimestamp() time.Time

	// the random message ID associated with this message
	GetID() util.Key

	// any additional arguments the command needs
	GetArgs() map[string]interface{}
}

type NodeMessage struct {
	Source  util.Key `codec:"source" json:"source"`
	Desired util.Key `codec:"desired" json:"desired"`

	// the message identifier
	ID util.Key `codec:"id" json:"id"`

	Command   string    `codec:"command" json:"command"`
	Timestamp time.Time `codec:"timestamp" json:"timestamp"`

	Args map[string]interface{} `codec:"args" json:"args"`
}

var (
	// MessageFormat is the codec format we'll use for all our communication
	MessageFormat = codec.MsgpackHandle{}
)

// NewMessage generates a new message object from the passed in parameters
func NewMessage(Source, Desired util.Key, Command string, Args map[string]interface{}) Message {
	m := &NodeMessage{
		Source:    Source,
		Desired:   Desired,
		Command:   Command,
		Timestamp: time.Now(),
		Args:      Args,
		ID:        util.NewKey(),
	}

	return m
}

func (n *NodeMessage) GetSourceKey() util.Key {
	return n.Source
}

func (n *NodeMessage) GetDesiredKey() util.Key {
	return n.Desired
}

func (n *NodeMessage) GetCommand() string {
	return n.Command
}

func (n *NodeMessage) GetTimestamp() time.Time {
	return n.Timestamp
}

func (n *NodeMessage) GetArgs() map[string]interface{} {
	return n.Args
}

func (n *NodeMessage) GetID() util.Key {
	return n.ID
}
