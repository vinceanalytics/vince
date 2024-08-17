package license

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/pkg/errors"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:embed PUBLIC_KEY
var publicKey []byte

func Verify(key []byte) (*v1.License, error) {
	ring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader(publicKey))
	if err != nil {
		return nil, err
	}
	b, err := armor.Decode(bytes.NewReader(key))
	if err != nil {
		return nil, fmt.Errorf("decoding license file %w", err)
	}
	md, err := openpgp.ReadMessage(b.Body, ring, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("reading PGP message from license file %w", err)
	}

	// We need to read the body for the signature verification check to happen.
	// md.Signature would be non-nil after reading the body if the verification is successful.
	buf, err := io.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, fmt.Errorf(" reading body from signed license file%w", err)
	}
	// This could be nil even if signature verification failed, so we also check Signature == nil
	// below.
	if md.SignatureError != nil {
		return nil, fmt.Errorf("verify license file %w", err)
	}
	if md.Signature == nil {
		return nil, errors.New("invalid signature while trying to verify license file")
	}
	var ls v1.License
	err = proto.Unmarshal(buf, &ls)
	if err != nil {
		return nil, fmt.Errorf("decoding license file data%w", err)
	}
	var set int
	ls.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		set++
		return true
	})
	if set != ls.ProtoReflect().Descriptor().Fields().Len() {
		return nil, errors.New("invalid license data")
	}
	return &ls, nil
}
