package code_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/scgolang/code"
)

type testHandler struct {
	t       *testing.T
	btnChan chan code.Button
	encChan chan code.Encoder
}

func newTestHandler(t *testing.T) testHandler {
	return testHandler{
		t:       t,
		btnChan: make(chan code.Button),
		encChan: make(chan code.Encoder),
	}
}

func (th testHandler) Button(btn code.Button) error {
	th.btnChan <- btn
	return nil
}

func (th testHandler) Encoder(enc code.Encoder) error {
	th.encChan <- enc
	return nil
}

func (th testHandler) Err(err error) {
	th.t.Fatal(err)
}

func TestCode(t *testing.T) {
	rand.Seed(time.Now().Unix())

	cd, err := code.New()
	if err != nil {
		t.SkipNow()
	}
	th := newTestHandler(t)
	cd.AddHandler(th)

	go func() {
		for enc := range th.encChan {
			fmt.Printf("enc %#v\n", enc)
		}
	}()

	for btn := range th.btnChan {
		fmt.Printf("btn %#v\n", btn)
	}
}
