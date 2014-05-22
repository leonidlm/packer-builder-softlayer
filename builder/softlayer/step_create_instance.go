package softlayer

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateInstance struct{}

func (self *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)

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
		ProvisioningSshKeyId: state.Get("ssh_key_id").(float64),
		BaseImageId:          config.BaseImageId,
		BaseOsCode:           config.BaseOsCode,
	}
	instanceData, _ := client.CreateInstance(*instanceDefinition)
	state.Put("instance_data", instanceData)

	ui.Say(fmt.Sprintf("Created instance '%s'", instanceData["globalIdentifier"].(string)))

	return multistep.ActionContinue
}

func (self *stepCreateInstance) Cleanup(state multistep.StateBag) {
}
