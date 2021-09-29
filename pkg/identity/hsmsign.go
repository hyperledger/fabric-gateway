// +build pkcs11

/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"

	"github.com/miekg/pkcs11"
)

// HSMSignerOptions are the options required for HSM Login.
type HSMSignerOptions struct {
	Label      string
	Pin        string
	Identifier string
	UserType   int
}

// HSMSignerFactory represents a factory to create HSM Signers.
type HSMSignerFactory struct {
	ctx *pkcs11.Ctx
}

// HSMSignClose Closes a HSM Sign.
type HSMSignClose = func() error

// NewHSMSignerFactory creates a new HSMSignerFactory. You only want one of these.
func NewHSMSignerFactory(library string) (*HSMSignerFactory, error) {
	if library == "" {
		return nil, fmt.Errorf("library path not provided")
	}

	ctx := pkcs11.New(library)
	if ctx == nil {
		return nil, fmt.Errorf("instantiation failed for %s", library)
	}

	if err := ctx.Initialize(); err != nil {
		return nil, fmt.Errorf("initialize failed: %w", err)
	}

	return &HSMSignerFactory{ctx}, nil
}

// NewHSMSigner creates a new HSM Signer. These are not Go Routine safe, do not share these across Go Routines.
func (factory *HSMSignerFactory) NewHSMSigner(options HSMSignerOptions) (Sign, HSMSignClose, error) {
	if options.Label == "" {
		return nil, nil, fmt.Errorf("no Label provided")
	}

	if options.Pin == "" {
		return nil, nil, fmt.Errorf("no Pin provided")
	}

	if options.Identifier == "" {
		return nil, nil, fmt.Errorf("no Identifier provided")
	}

	slots, err := factory.ctx.GetSlotList(true)
	if err != nil {
		return nil, nil, fmt.Errorf("get slot list failed: %w", err)
	}

	for _, slot := range slots {
		tokenInfo, err := factory.ctx.GetTokenInfo(slot)
		if err != nil || options.Label != tokenInfo.Label {
			continue
		}

		session, err := factory.createSession(slot, options.Pin)
		if err != nil {
			return nil, nil, err
		}

		privateKeyHandle, err := factory.findObjectInHSM(session, pkcs11.CKO_PRIVATE_KEY, options.Identifier)
		if err != nil {
			factory.ctx.CloseSession(session)
			return nil, nil, err
		}

		signer := &hsmSigner{
			ctx:              factory.ctx,
			session:          session,
			privateKeyHandle: privateKeyHandle,
		}
		return signer.Sign, signer.Close, nil
	}

	return nil, nil, fmt.Errorf("could not find token with label %s", options.Label)

}

// Dispose disposes of the HSMSignerFactory.
func (factory *HSMSignerFactory) Dispose() {
	factory.ctx.Finalize()
}

func (factory *HSMSignerFactory) findObjectInHSM(session pkcs11.SessionHandle, keyType uint, identifier string) (pkcs11.ObjectHandle, error) {
	template := []*pkcs11.Attribute{
		pkcs11.NewAttribute(pkcs11.CKA_CLASS, keyType),
		pkcs11.NewAttribute(pkcs11.CKA_ID, identifier),
	}
	if err := factory.ctx.FindObjectsInit(session, template); err != nil {
		return 0, fmt.Errorf("findObjectsInit failed: %w", err)
	}
	defer factory.ctx.FindObjectsFinal(session)

	// single session instance, assume one hit only
	objs, _, err := factory.ctx.FindObjects(session, 1)
	if err != nil {
		return 0, fmt.Errorf("findObjects failed: %w", err)
	}

	if len(objs) == 0 {
		return 0, fmt.Errorf("HSM Object not found for key [%s]", hex.EncodeToString([]byte(identifier)))
	}

	return objs[0], nil
}

func (factory *HSMSignerFactory) createSession(slot uint, pin string) (pkcs11.SessionHandle, error) {
	session, err := factory.ctx.OpenSession(slot, pkcs11.CKF_SERIAL_SESSION)
	if err != nil {
		return 0, fmt.Errorf("open session failed: %w", err)
	}

	if err := factory.ctx.Login(session, pkcs11.CKU_USER, pin); err != nil && err != pkcs11.Error(pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
		factory.ctx.CloseSession(session)
		return 0, fmt.Errorf("login failed: %w", err)
	}

	return session, nil
}

type hsmSigner struct {
	ctx              *pkcs11.Ctx
	lock             sync.Mutex
	session          pkcs11.SessionHandle
	privateKeyHandle pkcs11.ObjectHandle
}

func (signer *hsmSigner) Close() error {
	signer.lock.Lock()
	defer signer.lock.Unlock()

	return signer.ctx.CloseSession(signer.session)
}

func (signer *hsmSigner) Sign(digest []byte) ([]byte, error) {
	signature, err := signer.hsmSign(digest)
	if err != nil {
		return nil, err
	}

	r, s := new(big.Int), new(big.Int)
	sIndex := len(signature) / 2
	r.SetBytes(signature[0:sIndex])
	s.SetBytes(signature[sIndex:])

	// Only Elliptic of 256 byte keys are supported
	s, err = toLowSByCurve(elliptic.P256(), s)
	if err != nil {
		return nil, err
	}

	return marshalECDSASignature(r, s)
}

func (signer *hsmSigner) hsmSign(digest []byte) ([]byte, error) {
	signer.lock.Lock()
	defer signer.lock.Unlock()

	if err := signer.ctx.SignInit(
		signer.session,
		[]*pkcs11.Mechanism{pkcs11.NewMechanism(pkcs11.CKM_ECDSA, nil)},
		signer.privateKeyHandle,
	); err != nil {
		return nil, fmt.Errorf("sign initialize failed: %w", err)
	}

	signature, err := signer.ctx.Sign(signer.session, digest)
	if err != nil {
		return nil, fmt.Errorf("sign failed: %w", err)
	}

	return signature, nil
}
