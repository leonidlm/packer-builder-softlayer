package softlayer

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strconv"
	"strings"
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

	if config.ImageType == "standard" {
		ui.Say(fmt.Sprintf("Getting block devices for instance (id=%s)", instanceId))

		blockDevices, err := client.getBlockDevices(instanceId)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Unable to get list of block devices. Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		blockDeviceIds := make([]int64, len(blockDevices))
		diskCount := 0
		for _, val := range blockDevices {
			blockDevice := val.(map[string]interface{})
			diskImage := blockDevice["diskImage"].(map[string]interface{})
			name := diskImage["name"].(string)
			id := int64(blockDevice["id"].(float64))
			ui.Say(fmt.Sprintf("Found block device: n=%s, id=%v, diskImageId=%s, name=%s", blockDevice["device"], id, strconv.Itoa(int(blockDevice["diskImageId"].(float64))), name))

			if strings.Contains(name, "SWAP") {
				ui.Say("  ^--- Skip the above swap device ---^")
			} else {
				blockDeviceIds[diskCount] = id
				diskCount++
			}
		}
		blockDeviceIds = blockDeviceIds[:diskCount]
		ui.Say(fmt.Sprintf("Will caputure standard image using these disk images: %v", blockDeviceIds))

		_, err = client.captureStandardImage(instanceId, config.ImageName, config.ImageDescription, blockDeviceIds)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Say("Waiting for a standard image to finish it's creation...")
	} else {
		data, err := client.captureImage(instanceId, config.ImageName, config.ImageDescription)
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		imageId = data["globalIdentifier"].(string)
		state.Put("image_id", imageId)
		ui.Say(fmt.Sprintf("Waiting for a flex image (%s) to finish it's creation...", imageId))
	}

	// We are waiting for the instance since the waiting process checks for active transactions.
	// The image will be ready when no active transactions will be set for the snapshotted instance.
	err := client.waitForInstanceReady(instanceId, config.StateTimeout)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to become ACTIVE again after image creation call. Error: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if config.ImageType == "standard" {
		// Find the image id by listing all images and finding the one with the matching name.
		images, err := client.getBlockDeviceTemplateGroups()
		if err != nil {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). Unable to get list of images. Error: %s", instanceId, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		for _, val := range images {
			image := val.(map[string]interface{})
			ui.Say(fmt.Sprintf("Considering image: '%s'...", image["name"]))
			if image["name"] == config.ImageName && image["globalIdentifier"] != nil {
				imageId = image["globalIdentifier"].(string)
				ui.Say(fmt.Sprintf("Got the image id: %s", imageId))
				state.Put("image_id", imageId)
				break
			}
		}

		if imageId == "" {
			err := fmt.Errorf("Error while trying to capture an image from instance (id=%s). No image found with name '%s'.", instanceId, config.ImageName)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (self *stepCaptureImage) Cleanup(state multistep.StateBag) {
}
