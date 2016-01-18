package softlayer

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	client := state.Get("client").(*SoftlayerClient)
	instance := state.Get("instance_data").(map[string]interface{})
	instanceId := instance["globalIdentifier"].(string)
	ipAddress, err := client.getInstancePublicIp(instanceId)
	if err != nil {
		err := errors.New(fmt.Sprintf("Failed to fetch Public IP address for instance '%s'", instanceId))
		return "", err
	}

	return ipAddress, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(Config)
	privateKey := state.Get("ssh_private_key").(string)

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &ssh.ClientConfig{
		User: config.Comm.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}
