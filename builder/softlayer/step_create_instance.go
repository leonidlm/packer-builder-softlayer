package softlayer

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateInstance struct {
	instanceId string
}

func (self *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(*Config)
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
	ui.Say(fmt.Sprintf("Instance created: %q", instanceData["globalIdentifier"].(string)))

	return multistep.ActionContinue
}

func (self *stepCreateInstance) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if self.instanceId == "" {
		return
	}

	ui.Say(fmt.Sprintf("Waiting for the instance %q to have no active transactions before destroying it...", self.instanceId))

	// We should wait until the instance is up/have no transactions,
	// since if the instance will have some assigned transactions the destroy API call will fail
	err := client.waitForInstanceReady(self.instanceId, config.StateTimeout)
	if err != nil {
		log.Printf("Error destroying instance %q: %s", self.instanceId, err.Error())
		ui.Error(fmt.Sprintf("Error waiting for instance to become ACTIVE for instance %q", self.instanceId))
	}

	ui.Say(fmt.Sprintf("Destroying instance %q...", self.instanceId))
	err = client.DestroyInstance(self.instanceId)
	if err != nil {
		log.Printf("Error destroying instance %q: %s", self.instanceId, err)
		ui.Error(fmt.Sprintf("Error cleaning up the instance. Please delete the instance %q manually", self.instanceId))
	}
	ui.Say(fmt.Sprintf("Instance %q destroyed", self.instanceId))
}
