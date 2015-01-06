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
	instance := state.Get("instance_data").(map[string]interface{})
	config := state.Get("config").(config)
	instanceId := instance["globalIdentifier"].(string)
	var imageId string

	ui.Say(fmt.Sprintf("Preparing for capturing the instance image. Image snapshot type is '%s'.", config.ImageType))

	if config.ImageType == IMAGE_TYPE_STANDARD {
		ui.Say(fmt.Sprintf("Getting block devices for instance (id=%s)", instanceId))

		blockDevices, err := client.getBlockDevices(instanceId)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Unable to get list of block devices. Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		blockDeviceIds := client.findNonSwapBlockDeviceIds(blockDevices)
		ui.Say(fmt.Sprintf("Will caputure standard image using these block devices: %v", blockDeviceIds))

		_, err = client.captureStandardImage(instanceId, config.ImageName, config.ImageDescription, blockDeviceIds)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		imageId, err = client.findImageIdByName(config.ImageName)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Could not get image id. Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	} else {
		// Flex Image
		data, err := client.captureImage(instanceId, config.ImageName, config.ImageDescription)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		imageId = data["globalIdentifier"].(string)
	}

	state.Put("image_id", imageId)
	ui.Say(fmt.Sprintf("Waiting for image (%s) to finish its creation...", imageId))

	// We are waiting for the instance since the waiting process checks for active transactions.
	// The image will be ready when no active transactions will be set for the snapshotted instance.
	err := client.waitForInstanceReady(instanceId, config.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to become ACTIVE again after image creation call. Error: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (self *stepCaptureImage) Cleanup(state multistep.StateBag) {
}
