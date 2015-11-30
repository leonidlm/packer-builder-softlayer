package softlayer

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepWaitforInstance struct{}

func (self *stepWaitforInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance_data").(map[string]interface{})
	id := instance["globalIdentifier"].(string)

	ui.Say(fmt.Sprintf("Waiting for the instance %q to become ACTIVE...", id))

	err := client.waitForInstanceReady(id, config.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance %q to become ACTIVE: %s", id, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Instance %q is ACTIVE", id))
	return multistep.ActionContinue
}

func (self *stepWaitforInstance) Cleanup(state multistep.StateBag) {}
