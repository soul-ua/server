package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Envelope struct {
	From        string
	To          string
	PayloadType string // todo: type of Payload can be encrypted as part of Payload
	Payload     []byte
}

type envelopeWire struct {
	From        string
	To          string
	PayloadType string
	Payload     []byte
}

func (e *Envelope) Pack() ([]byte, error) {
	// todo: this is golang specific, should be replaced with protobuf or other common binary format
	packed := bytes.Buffer{}
	enc := gob.NewEncoder(&packed)
	err := enc.Encode(envelopeWire{
		From:        e.From,
		To:          e.To,
		PayloadType: e.PayloadType,
		Payload:     e.Payload,
	})
	if err != nil {
		return nil, fmt.Errorf("failed To encode wire envelope: %w", err)
	}

	return packed.Bytes(), nil
}

func UnpackEnvelope(packed []byte) (*Envelope, error) {
	packetBuf := bytes.NewBuffer(packed)

	dec := gob.NewDecoder(packetBuf)
	var wire envelopeWire
	err := dec.Decode(&wire)
	if err != nil {
		return nil, fmt.Errorf("failed To decode wire envelope: %w", err)
	}

	return &Envelope{
		From:        wire.From,
		To:          wire.To,
		PayloadType: wire.PayloadType,
		Payload:     wire.Payload,
	}, nil
}
