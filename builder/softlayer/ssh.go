package softlayer

import (
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
)

func sshAddress(state multistep.StateBag) (string, error) {
	config := state.Get("config").(config)
	client := state.Get("client").(*SoftlayerClient)
	instance := state.Get("instance_data").(map[string]interface{})
	instanceId := instance["globalIdentifier"].(string)
	ipAddress, err := client.getInstancePublicIp(instanceId)
	if err != nil {
		err := errors.New(fmt.Sprintf("Failed to fetch Public IP address for instance '%s'", instanceId))
		return "", err
	}

	return fmt.Sprintf("%s:%d", ipAddress, config.SshPort), nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	config := state.Get("config").(config)
	privateKey := state.Get("ssh_private_key").(string)

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	return &ssh.ClientConfig{
		User: config.SshUserName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}, nil
}
