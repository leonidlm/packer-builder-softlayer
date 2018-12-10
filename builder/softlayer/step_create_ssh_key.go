package softlayer

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSshKey struct {
	keyId          int64
	PrivateKeyFile []byte
}

func (self *stepCreateSshKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if bytes.Compare(self.PrivateKeyFile, []byte("")) != 0 {
		ui.Say(fmt.Sprintf("Reading private key file (%s)...", self.PrivateKeyFile))

		privateKeyBytes, err := ioutil.ReadFile(string(self.PrivateKeyFile))
		if err != nil {
			state.Put("error", fmt.Errorf("Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}

		state.Put("ssh_private_key", string(privateKeyBytes))

		return multistep.ActionContinue
	}

	client := state.Get("client").(*SoftlayerClient)
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

	ui.Say(fmt.Sprintf("Created SSH key with id '%d'", keyId))

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
		ui.Error(fmt.Sprintf("Error cleaning up ssh key. Please delete the key (%d) manually", self.keyId))
	}
}
