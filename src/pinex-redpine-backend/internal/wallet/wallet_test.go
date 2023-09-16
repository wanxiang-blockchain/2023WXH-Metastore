package wallet

import (
	"fmt"
	"testing"
)

var privatePath = "./private"

// var data = "Hello world"
var data = "Welcome to PINEX"

//var data = "Welcome to PINEX!\n\nThis request will not trigger a blockchain transaction or cost any gas fees. It is only used to authorise logging into PINEX.\n\nYour authentication status will reset after 12 hours.\n\nWallet address:\n0x4d2104b5027264e622dba567c4fdcdad424e26ad\n\nNonce:\n540619"

func TestNewWallet(t *testing.T) {
	w, err := NewWallet(privatePath)
	if err != nil {
		t.Fatalf("NewWallet err %s", err.Error())
	}
	fmt.Printf("wallet address: %v\n", w.address)
}

func TestLoadWallet(t *testing.T) {
	w, err := LoadWallet(privatePath)
	if err != nil {
		t.Fatalf("NewWallet err %s", err.Error())
	}
	fmt.Printf("wallet address: %v\n", w.address)
}

func TestSign(t *testing.T) {
	w, err := LoadWallet(privatePath)
	if err != nil {
		t.Fatalf("NewWallet err %s", err.Error())
	}
	fmt.Printf("wallet address: %v\n", w.address)

	signature, err := w.Sign(data)
	if err != nil {
		t.Fatalf("Sign err %s", err.Error())
	}
	fmt.Printf("len(signature): %d\n", len(signature))
	fmt.Printf("signature: %v\n", signature)
	fmt.Printf("signature string: %x\n", string(signature))
}

func TestVerifySign(t *testing.T) {
	w, err := LoadWallet(privatePath)
	if err != nil {
		t.Fatalf("NewWallet err %s", err.Error())
	}
	fmt.Printf("wallet address: %v\n", w.address)

	signature, err := w.Sign(data)
	if err != nil {
		t.Fatalf("Sign err %s", err.Error())
	}
	fmt.Printf("len(signature): %d\n", len(signature))
	fmt.Printf("signature: %v\n", signature)
	fmt.Printf("signature string: %x\n", string(signature))

	b, err := w.Verify(data, signature)
	if err != nil {
		t.Fatalf("Verify err %s", err.Error())
	}

	if !b {
		t.Fatalf("Verify invalid")
	}
}

func TestEthVerifySign(t *testing.T) {
	w, err := LoadWallet(privatePath)
	if err != nil {
		t.Fatalf("NewWallet err %s", err.Error())
	}
	fmt.Printf("wallet address: %v\n", w.address)

	signature, err := w.EthSign(data)
	if err != nil {
		t.Fatalf("Sign err %s", err.Error())
	}
	fmt.Printf("len(signature): %d\n", len(signature))
	fmt.Printf("signature: %v\n", signature)
	fmt.Printf("signature string: %x\n", string(signature))

	b, err := w.EthVerify(data, signature)
	if err != nil {
		t.Fatalf("Verify err %s", err.Error())
	}

	if !b {
		t.Fatalf("Verify invalid")
	}
}
