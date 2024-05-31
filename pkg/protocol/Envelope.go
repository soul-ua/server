package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Envelope struct {
	ID          string
	From        string
	To          string
	Time        int64
	PayloadType string // todo: type of Payload can be encrypted as part of Payload
	Payload     []byte
}

type envelopeWire struct {
	ID          string
	From        string
	To          string
	Time        int64
	PayloadType string
	Payload     []byte
}

func (e *Envelope) Pack() ([]byte, error) {
	// todo: this is golang specific, should be replaced with protobuf or other common binary format
	packed := bytes.Buffer{}
	enc := gob.NewEncoder(&packed)
	err := enc.Encode(envelopeWire{
		ID:          e.ID,
		From:        e.From,
		To:          e.To,
		Time:        e.Time,
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
		ID:          wire.ID,
		From:        wire.From,
		To:          wire.To,
		Time:        wire.Time,
		PayloadType: wire.PayloadType,
		Payload:     wire.Payload,
	}, nil
}
