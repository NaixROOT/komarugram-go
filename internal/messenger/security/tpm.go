// SPDX-License-Identifier: Unlicense OR MIT

package security

import (
	"fmt"

	"github.com/google/go-tpm/legacy/tpm2"
)

type hardwareTPM struct{}

var storageRootTemplate = tpm2.Public{
	Type:       tpm2.AlgECC,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagSensitiveDataOrigin | tpm2.FlagUserWithAuth | tpm2.FlagRestricted | tpm2.FlagDecrypt | tpm2.FlagNoDA,
	ECCParameters: &tpm2.ECCParams{
		Symmetric: &tpm2.SymScheme{Alg: tpm2.AlgAES, KeyBits: 128, Mode: tpm2.AlgCFB},
		CurveID:   tpm2.CurveNISTP256,
		Point: tpm2.ECPoint{
			XRaw: make([]byte, 32),
			YRaw: make([]byte, 32),
		},
	},
}

var sealedObjectTemplate = tpm2.Public{
	Type:       tpm2.AlgKeyedHash,
	NameAlg:    tpm2.AlgSHA256,
	Attributes: tpm2.FlagFixedTPM | tpm2.FlagFixedParent | tpm2.FlagUserWithAuth,
}

func (hardwareTPM) Probe() error {
	rw, err := tpm2.OpenTPM()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	return rw.Close()
}

func (hardwareTPM) Seal(secret, authorization []byte) (public, private []byte, retErr error) {
	rw, err := tpm2.OpenTPM()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() {
		if err := rw.Close(); retErr == nil && err != nil {
			retErr = err
		}
	}()
	parent, _, err := tpm2.CreatePrimary(rw, tpm2.HandleOwner, tpm2.PCRSelection{}, "", "", storageRootTemplate)
	if err != nil {
		return nil, nil, fmt.Errorf("create storage root: %w", err)
	}
	defer tpm2.FlushContext(rw, parent)
	private, public, _, _, _, err = tpm2.CreateKeyWithSensitive(rw, parent, tpm2.PCRSelection{}, "", string(authorization), sealedObjectTemplate, secret)
	if err != nil {
		return nil, nil, fmt.Errorf("create sealed object: %w", err)
	}
	return public, private, nil
}

func (hardwareTPM) Unseal(public, private, authorization []byte) (secret []byte, retErr error) {
	rw, err := tpm2.OpenTPM()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUnavailable, err)
	}
	defer func() {
		if err := rw.Close(); retErr == nil && err != nil {
			retErr = err
		}
	}()
	parent, _, err := tpm2.CreatePrimary(rw, tpm2.HandleOwner, tpm2.PCRSelection{}, "", "", storageRootTemplate)
	if err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	defer tpm2.FlushContext(rw, parent)
	object, _, err := tpm2.Load(rw, parent, "", public, private)
	if err != nil {
		return nil, fmt.Errorf("load sealed object: %w", err)
	}
	defer tpm2.FlushContext(rw, object)
	secret, err = tpm2.Unseal(rw, object, string(authorization))
	if err != nil {
		return nil, err
	}
	return secret, nil
}
