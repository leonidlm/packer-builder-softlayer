package softlayer

import (
	"code.google.com/p/gosshold/ssh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"strings"
)

type stepCreateSshKey struct {
	keyId float64
}

func (self *stepCreateSshKey) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary ssh key for the instance...")

	rsaKey, _ := rsa.GenerateKey(rand.Reader, 2014)

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(rsaKey)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Set the private key in the statebag for later
	state.Put("ssh_private_key", string(pem.EncodeToMemory(&privBlk)))

	pub, _ := ssh.NewPublicKey(&rsaKey.PublicKey)
	publicKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))

	// The name of the public key on DO
	label := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	keyId, _ := client.UploadSshKey(label, publicKey)

	self.keyId = keyId
	state.Put("ssh_key_id", keyId)

	ui.Say(fmt.Sprintf("Created SSH key with id '%i'", keyId))

	return multistep.ActionContinue
}

func (self *stepCreateSshKey) Cleanup(state multistep.StateBag) {
	// use the self.keyId to cleanup the temporary ssh key
}
