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
	"log"
	"strings"
)

type stepCreateSshKey struct {
	keyId int64
}

func (self *stepCreateSshKey) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary ssh key for the instance...")

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(rsaKey)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Set the private key in the statebag for later
	state.Put("ssh_private_key", string(pem.EncodeToMemory(&privBlk)))

	pub, err := ssh.NewPublicKey(&rsaKey.PublicKey)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	publicKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))

	// The name of the public key
	label := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	keyId, err := client.UploadSshKey(label, publicKey)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	self.keyId = keyId
	state.Put("ssh_key_id", keyId)

	ui.Say(fmt.Sprintf("Created SSH key with id '%i'", keyId))

	return multistep.ActionContinue
}

func (self *stepCreateSshKey) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if self.keyId == 0 {
		return
	}

	client := state.Get("client").(*SoftlayerClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	err := client.DestroySshKey(self.keyId)

	if err != nil {
		log.Printf("Error cleaning up ssh key: %v", err.Error())
		ui.Error(fmt.Sprintf("Error cleaning up ssh key. Please delete the key (%i) manually", self.keyId))
	}
}
