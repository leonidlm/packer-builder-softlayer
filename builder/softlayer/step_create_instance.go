package softlayer

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateInstance struct {
	instanceId string
}

func (self *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// The ssh_key_id can be empty if the user specified a private key
	sshKeyId := state.Get("ssh_key_id")
	var ProvisioningSshKeyId int64 = 0
	if sshKeyId != nil {
		ProvisioningSshKeyId = sshKeyId.(int64)
	}

	instanceDefinition := &InstanceType{
		HostName:             config.InstanceName,
		Domain:               config.InstanceDomain,
		Datacenter:           config.DatacenterName,
		Cpus:                 config.InstanceCpu,
		Memory:               config.InstanceMemory,
		HourlyBillingFlag:    true,
		LocalDiskFlag:        true,
		DiskCapacity:         config.InstanceDiskCapacity,
		NetworkSpeed:         config.InstanceNetworkSpeed,
		ProvisioningSshKeyId: ProvisioningSshKeyId,
		BaseImageId:          config.BaseImageId,
		BaseOsCode:           config.BaseOsCode,
	}

	ui.Say("Creating an instance...")
	instanceData, err := client.CreateInstance(*instanceDefinition)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_data", instanceData)
	self.instanceId = instanceData["globalIdentifier"].(string)
	ui.Say(fmt.Sprintf("Created instance, id: '%s'", instanceData["globalIdentifier"].(string)))

	return multistep.ActionContinue
}

func (self *stepCreateInstance) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	if self.instanceId == "" {
		return
	}

	ui.Say("Waiting for the instance to have no active transactions before destroying it...")

	// We should wait until the instance is up/have no transactions,
	// since if the instance will have some assigned transactions the destroy API call will fail
	err := client.waitForInstanceReady(self.instanceId, config.StateTimeout)
	if err != nil {
		log.Printf("Error destroying instance: %v", err.Error())
		ui.Error(fmt.Sprintf("Error waiting for instance to become ACTIVE for instance (%s)", self.instanceId))
	}

	ui.Say("Destroying instance...")
	err = client.DestroyInstance(self.instanceId)
	if err != nil {
		log.Printf("Error destroying instance: %v", err.Error())
		ui.Error(fmt.Sprintf("Error cleaning up the instance. Please delete the instance (%s) manually", self.instanceId))
	}
}
