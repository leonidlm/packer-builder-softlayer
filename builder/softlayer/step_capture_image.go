package softlayer

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCaptureImage struct{}

func (self *stepCaptureImage) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Preparing for capturing the instance (flex) image")

	instance := state.Get("instance_data").(map[string]interface{})
	instanceId := instance["globalIdentifier"].(string)
	config := state.Get("config").(config)

	data, err := client.captureImage(instanceId, config.ImageName, config.ImageDescription)
	if err != nil {
		err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Error: %s", instanceId, err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imageId := data["globalIdentifier"]
	ui.Say(fmt.Sprintf("Waiting for an image (%s) to finish it's creation...", imageId))

	// We are waiting for the instance since the waiting process checks for active transactions.
	// The image will be ready when no active transactions will be set for the snapshotted instance.
	err = client.waitForInstanceReady(instanceId, config.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to become ACTIVE again after image creation call. Error: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (self *stepCaptureImage) Cleanup(state multistep.StateBag) {
}
