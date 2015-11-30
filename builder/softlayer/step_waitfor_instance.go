package softlayer

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepWaitforInstance struct{}

func (self *stepWaitforInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Waiting for the instance to become ACTIVE...")

	instance := state.Get("instance_data").(map[string]interface{})
	err := client.waitForInstanceReady(instance["globalIdentifier"].(string), config.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to become ACTIVE: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (self *stepWaitforInstance) Cleanup(state multistep.StateBag) {
}
