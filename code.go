// Package code is a Go package for talking to the Livid Code v2 MIDI controller.
package code

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/scgolang/midi"
)

const (
	Controls = "Controls"
)

// Code represents a connection to a Livid Code v2.
type Code struct {
	*midi.Device

	handlers chan<- Handler
}

// Button represents one of the buttons on the Code.
// The buttons are numbered 1 to 13 with 1 being the upper left button,
// 5 being the bottom left button, and 13 being the bottom right button.
type Button struct {
	Index int
	Value int
}

// Encoder represents a value of one of the encoders.
// The encoders are numbered 1 to 32 with 1 being the upper left encoder,
// 4 being the bottom left encoder and so on horizontally across the controller,
// with the bottom right encoder being 32.
type Encoder struct {
	Index int
	Value int
}

// Handler provides an easy way to receive MIDI events from the Code.
type Handler interface {
	Button(Button) error
	Encoder(Encoder) error
	Err(error)
}

// New creates a new connection to the device.
func New() (*Code, error) {
	devices, err := midi.Devices()
	if err != nil {
		return nil, errors.Wrap(err, "listing devices")
	}
	var (
		hch = make(chan Handler, 1)
		c   = &Code{handlers: hch}
	)
	for _, d := range devices {
		if d.Name == Controls {
			c.Device = d
		}
	}
	if c.Device == nil {
		return nil, errors.New("finding device")
	}
	if err := c.Open(); err != nil {
		return nil, errors.Wrap(err, "opening device")
	}
	pkts, err := c.Packets()
	if err != nil {
		return nil, errors.Wrap(err, "getting packets channel")
	}
	go handlers(pkts, hch)

	return c, nil
}

// AddHandler adds a handler that will receive all the events coming from the device.
func (c *Code) AddHandler(h Handler) {
	c.handlers <- h
}

// SetButton sets the value of a button on the Code.
func (c *Code) SetButton(b Button) error {
	return nil
}

// SetEncoder sets the value of an encoder on the Code.
func (c *Code) SetEncoder(e Encoder) error {
	return nil
}

func handlers(pkts <-chan midi.Packet, hch <-chan Handler) {
	handlers := []Handler{}

	for {
		select {
		case pkt := <-pkts:
			if err := handlePacket(handlers, pkt); err != nil {
				// TODO: handle error
				fmt.Fprintf(os.Stderr, "%s (packet %#v)\n", err, pkt.Data)
			}
		case h := <-hch:
			handlers = append(handlers, h)
		}
	}
}

func handlePacket(handlers []Handler, p midi.Packet) error {
	if p.Err != nil {
		return p.Err // An error reading data from the device.
	}
	switch p.Data[0] {
	case midi.CC:
		var (
			enc  = Encoder{Index: int(p.Data[1]), Value: int(p.Data[2])}
			errs []string
		)
		for _, h := range handlers {
			if err := h.Encoder(enc); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return errors.New(strings.Join(errs, ", and "))
	case midi.Note:
		var (
			btn  = Button{Index: int(p.Data[1]) - 32, Value: int(p.Data[2])}
			errs []string
		)
		for _, h := range handlers {
			if err := h.Button(btn); err != nil {
				errs = append(errs, err.Error())
			}
		}
		if len(errs) == 0 {
			return nil
		}
		return errors.New(strings.Join(errs, ", and "))
	default:
		return errors.Errorf("unrecognized status byte: %x", p.Data[0])
	}
}
