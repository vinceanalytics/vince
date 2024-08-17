package main

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/openpgp/s2k"
)

var (
	name    = flag.String("name", "John Doe", "")
	comment = flag.String("comment", "Some comment", "")
	email   = flag.String("email", "acme@example.com", "")
)

func main() {
	flag.Parse()
	out := flag.Arg(0)
	uid := packet.NewUserId(*name, *comment, *email)
	if uid == nil {
		log.Fatal("nil uid")
	}
	sk, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	ek, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	creationTime := time.Now()
	e := &openpgp.Entity{
		PrimaryKey: packet.NewECDSAPublicKey(creationTime, &sk.PublicKey),
		PrivateKey: packet.NewECDSAPrivateKey(creationTime, sk),
		Identities: make(map[string]*openpgp.Identity),
	}
	isPrimaryId := true
	e.Identities[uid.Id] = &openpgp.Identity{
		Name:   uid.Id,
		UserId: uid,
		SelfSignature: &packet.Signature{
			CreationTime: creationTime,
			SigType:      packet.SigTypePositiveCert,
			PubKeyAlgo:   packet.PubKeyAlgoECDSA,
			Hash:         crypto.SHA256,
			IsPrimaryId:  &isPrimaryId,
			FlagsValid:   true,
			FlagSign:     true,
			FlagCertify:  true,
			IssuerKeyId:  &e.PrimaryKey.KeyId,
		},
	}
	err = e.Identities[uid.Id].SelfSignature.SignUserId(uid.Id, e.PrimaryKey, e.PrivateKey, nil)
	if err != nil {
		log.Fatal(err)
	}
	h, _ := s2k.HashToHashId(crypto.SHA256)
	e.Identities[uid.Id].SelfSignature.PreferredHash = []uint8{h}

	e.Subkeys = make([]openpgp.Subkey, 1)
	e.Subkeys[0] = openpgp.Subkey{
		PublicKey:  packet.NewECDSAPublicKey(creationTime, &ek.PublicKey),
		PrivateKey: packet.NewECDSAPrivateKey(creationTime, ek),
		Sig: &packet.Signature{
			CreationTime:              creationTime,
			SigType:                   packet.SigTypeSubkeyBinding,
			PubKeyAlgo:                packet.PubKeyAlgoECDSA,
			Hash:                      crypto.SHA256,
			FlagsValid:                true,
			FlagEncryptStorage:        true,
			FlagEncryptCommunications: true,
			IssuerKeyId:               &e.PrimaryKey.KeyId,
		},
	}
	e.Subkeys[0].PublicKey.IsSubkey = true
	e.Subkeys[0].PrivateKey.IsSubkey = true
	err = e.Subkeys[0].Sig.SignKey(e.Subkeys[0].PublicKey, e.PrivateKey, nil)
	if err != nil {
		log.Fatal(err)
	}
	var o bytes.Buffer
	w, err := armor.Encode(&o, openpgp.PublicKeyType, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = e.Serialize(w)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(out, "PUBLIC_KEY"), o.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
	o.Reset()
	w, err = armor.Encode(&o, openpgp.PrivateKeyType, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = e.SerializePrivate(w, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(out, "PRIVATE_KEY"), o.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
}
