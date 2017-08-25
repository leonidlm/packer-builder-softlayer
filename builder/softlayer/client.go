package softlayer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"strconv"
)

const SOFTLAYER_API_URL = "api.softlayer.com/rest/v3"

type SoftlayerClient struct {
	// The http client for communicating
	http *http.Client

	// Credentials
	user   string
	apiKey string
}

type SoftLayerRequest struct {
	Parameters interface{} `json:"parameters"`
}

// Based on: http://sldn.softlayer.com/reference/datatypes/SoftLayer_Container_Virtual_Guest_Configuration/
type InstanceType struct {
	HostName               string `json:"hostname"`
	Domain                 string
	Datacenter             string
	Cpus                   int
	Memory                 int64
	HourlyBillingFlag      bool
	LocalDiskFlag          bool
	PrivateNetworkOnlyFlag bool
	DiskCapacities         []int
	NetworkSpeed           int
	ProvisioningSshKeyId   int64
	BaseImageId            string
	BaseOsCode             string
}

type InstanceReq struct {
	HostName                 string                    `json:"hostname"`
	Domain                   string                    `json:"domain"`
	Datacenter               *Datacenter               `json:"datacenter"`
	Cpus                     int                       `json:"startCpus"`
	Memory                   int64                     `json:"maxMemory"`
	HourlyBillingFlag        bool                      `json:"hourlyBillingFlag"`
	LocalDiskFlag            bool                      `json:"localDiskFlag"`
	PrivateNetworkOnlyFlag   bool                      `json:"privateNetworkOnlyFlag"`
	NetworkComponents        []*NetworkComponent       `json:"networkComponents"`
	BlockDeviceTemplateGroup *BlockDeviceTemplateGroup `json:"blockDeviceTemplateGroup,omitempty"`
	BlockDevices             []*BlockDevice            `json:"blockDevices,omitempty"`
	OsReferenceCode          string                    `json:"operatingSystemReferenceCode,omitempty"`
	SshKeys                  []*SshKey                 `json:"sshKeys,omitempty"`
}

type InstanceImage struct {
	Descption string `json:"description"`
	Name      string `json:"name"`
	Summary   string `json:"summary"`
}

type Datacenter struct {
	Name string `json:"name"`
}

type NetworkComponent struct {
	MaxSpeed int `json:"maxSpeed"`
}

type BlockDeviceTemplateGroup struct {
	Id string `json:"globalIdentifier"`
}

type DiskImage struct {
	Capacity int `json:"capacity"`
}

type SshKey struct {
	Id    int64  `json:"id,omitempty"`
	Key   string `json:"key,omitempty"`
	Label string `json:"label,omitempty"`
}

type BlockDevice struct {
	Id        int64      `json:"id,omitempty"`
	Device    string     `json:"device,omitempty"`
	DiskImage *DiskImage `json:"diskImage,omitempty"`
}

func (self SoftlayerClient) New(user string, key string) *SoftlayerClient {
	return &SoftlayerClient{
		http: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
		user:   user,
		apiKey: key,
	}
}

func (self SoftlayerClient) generateRequestBody(params ...interface{}) (*bytes.Buffer, error) {
	softlayerRequest := &SoftLayerRequest{
		Parameters: params,
	}

	body, err := json.Marshal(softlayerRequest)
	if err != nil {
		return nil, err
	}

	log.Printf("Generated a request: %s", body)

	return bytes.NewBuffer(body), nil
}

func (self SoftlayerClient) hasErrors(body map[string]interface{}) error {
	if errString, ok := body["error"]; !ok {
		return nil
	} else {
		return errors.New(errString.(string))
	}
}

func (self SoftlayerClient) doRawHttpRequest(path string, requestType string, requestBody *bytes.Buffer) ([]byte, error) {
	url := fmt.Sprintf("https://%s:%s@%s/%s", self.user, self.apiKey, SOFTLAYER_API_URL, path)
	log.Printf("Sending new request to softlayer: %s", url)

	// Create the request object
	var lastResponse http.Response
	switch requestType {
	case "POST", "DELETE":
		req, err := http.NewRequest(requestType, url, requestBody)

		if err != nil {
			return nil, err
		}
		resp, err := self.http.Do(req)

		if err != nil {
			return nil, err
		} else {
			lastResponse = *resp
		}
	case "GET":
		resp, err := http.Get(url)

		if err != nil {
			return nil, err
		} else {
			lastResponse = *resp
		}
	default:
		return nil, errors.New(fmt.Sprintf("Undefined request type '%s', only GET/POST/DELETE are available!", requestType))
	}

	responseBody, err := ioutil.ReadAll(lastResponse.Body)
	lastResponse.Body.Close()
	if err != nil {
		return nil, err
	}

	log.Printf("Received response from SoftLayer: %s", responseBody)
	return responseBody, nil
}

func (self SoftlayerClient) doHttpRequest(path string, requestType string, requestBody *bytes.Buffer) ([]interface{}, error) {
	responseBody, err := self.doRawHttpRequest(path, requestType, requestBody)
	if err != nil {
		err := errors.New(fmt.Sprintf("Failed to get proper HTTP response from SoftLayer API"))
		return nil, err
	}

	var decodedResponse interface{}
	err = json.Unmarshal(responseBody, &decodedResponse)
	if err != nil {
		err := errors.New(fmt.Sprintf("Failed to decode JSON response from SoftLayer: %s | %s", responseBody, err))
		return nil, err
	}

	switch v := decodedResponse.(type) {
	case []interface{}:
		return v, nil
	case map[string]interface{}:
		if err := self.hasErrors(v); err != nil {
			return nil, err
		}

		return []interface{} {v,}, nil

	case nil:
		return []interface{} {nil,}, nil
	default:
		return nil, errors.New("Unexpected type in HTTP response")
	}
}

func (self SoftlayerClient) CreateInstance(instance InstanceType) (map[string]interface{}, error) {
	// SoftLayer API puts some limitations on hostname and domain fields of the request
	validName, err := regexp.Compile("[^A-Za-z0-9\\-\\.]+")
	if err != nil {
		return nil, err
	}

	instance.HostName = validName.ReplaceAllString(instance.HostName, "")
	instance.Domain = validName.ReplaceAllString(instance.Domain, "")

	// Construct the instance request object which will be decoded into json and posted to the API
	instanceRequest := &InstanceReq{
		HostName: instance.HostName,
		Domain:   instance.Domain,
		Datacenter: &Datacenter{
			Name: instance.Datacenter,
		},
		Cpus:              instance.Cpus,
		Memory:            instance.Memory,
		HourlyBillingFlag: true,
		PrivateNetworkOnlyFlag: instance.PrivateNetworkOnlyFlag,
		LocalDiskFlag:     false,
		NetworkComponents: []*NetworkComponent{
			&NetworkComponent{
				MaxSpeed: instance.NetworkSpeed,
			},
		},
	}

	if instance.ProvisioningSshKeyId != 0 {
		instanceRequest.SshKeys = []*SshKey{
			&SshKey{
				Id: instance.ProvisioningSshKeyId,
			},
		}
	}

	if instance.BaseImageId != "" {
		instanceRequest.BlockDeviceTemplateGroup = &BlockDeviceTemplateGroup{
			Id: instance.BaseImageId,
		}
	} else {
		instanceRequest.OsReferenceCode = instance.BaseOsCode

		for index, element := range instance.DiskCapacities {
			var device string // device 1 is reserved for swap, so we need to skip that
			if (index == 0) {
			  device= "0"
			} else {
			  device= strconv.Itoa(index + 1)
			}
			instanceRequest.BlockDevices = append(instanceRequest.BlockDevices, &BlockDevice{
				Device: device,
				DiskImage: &DiskImage{
					Capacity: element,
				}})
		}
	}

	requestBody, err := self.generateRequestBody(instanceRequest)
	if err != nil {
		return nil, err
	}

	data, err := self.doHttpRequest("SoftLayer_Virtual_Guest/createObject.json", "POST", requestBody)
	if err != nil {
		return nil, err
	}

	return data[0].(map[string]interface{}), err
}

func (self SoftlayerClient) DestroyInstance(instanceId string) error {
	response, err := self.doRawHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s.json", instanceId), "DELETE", new(bytes.Buffer))

	log.Printf("Deleted an Instance with id (%s), response: %s", instanceId, response)

	if res := string(response[:]); res != "true" {
		return errors.New(fmt.Sprintf("Failed to destroy and instance wit id '%s', got '%s' as response from the API.", instanceId, res))
	}

	return err
}

func (self SoftlayerClient) UploadSshKey(label string, publicKey string) (keyId int64, err error) {
	sshKeyRequest := &SshKey{
		Key:   publicKey,
		Label: label,
	}

	requestBody, err := self.generateRequestBody(sshKeyRequest)
	if err != nil {
		return 0, err
	}

	data, err := self.doHttpRequest("SoftLayer_Security_Ssh_Key/createObject.json", "POST", requestBody)
	if err != nil {
		return 0, err
	}

	return int64(data[0].(map[string]interface{})["id"].(float64)), err
}

func (self SoftlayerClient) DestroySshKey(keyId int64) error {
	response, err := self.doRawHttpRequest(fmt.Sprintf("SoftLayer_Security_Ssh_Key/%v.json", int(keyId)), "DELETE", new(bytes.Buffer))

	log.Printf("Deleted an SSH Key with id (%v), response: %s", keyId, response)
	if res := string(response[:]); res != "true" {
		return errors.New(fmt.Sprintf("Failed to destroy and SSH key wit id '%v', got '%s' as response from the API.", keyId, res))
	}

	return err
}

func (self SoftlayerClient) getInstancePublicIp(instanceId string) (string, error) {
	response, err := self.doRawHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/getPrimaryIpAddress.json", instanceId), "GET", nil)
	if err != nil {
		return "", nil
	}

	var validIp = regexp.MustCompile(`[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,4}`)
	ipAddress := validIp.Find(response)

	return string(ipAddress), nil
}

func (self SoftlayerClient) getInstancePrivateIp(instanceId string) (string, error) {
	response, err := self.doRawHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/getPrimaryBackendIpAddress.json", instanceId), "GET", nil)
	if err != nil {
		return "", nil
	}

	var validIp = regexp.MustCompile(`[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,4}\.[0-9]{1,4}`)
	ipAddress := validIp.Find(response)

	return string(ipAddress), nil
}

func (self SoftlayerClient) getBlockDevices(instanceId string) ([]interface{}, error) {
	data, err := self.doHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/getBlockDevices.json?objectMask=mask.diskImage.name", instanceId), "GET", nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (self SoftlayerClient) findNonSwapBlockDeviceIds(blockDevices []interface{}) ([]int64) {
	blockDeviceIds := make([]int64, len(blockDevices))
	deviceCount := 0

	for _, val := range blockDevices {
		blockDevice := val.(map[string]interface{})
		diskImage := blockDevice["diskImage"].(map[string]interface{})
		name := diskImage["name"].(string)
		id := int64(blockDevice["id"].(float64))

		// Skip both SWAP and METADATA devices
		// Reference - https://github.com/softlayer/softlayer-python/pull/776
		if ( !strings.Contains(name, "SWAP") && !strings.Contains(name, "METADATA") ) {
			blockDeviceIds[deviceCount] = id
			deviceCount++
		}
	}

	return blockDeviceIds[:deviceCount]
}

func (self SoftlayerClient) getBlockDeviceTemplateGroups() ([]interface{}, error) {
	data, err := self.doHttpRequest("SoftLayer_Account/getBlockDeviceTemplateGroups.json", "GET", nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (self SoftlayerClient) findImageIdByName(imageName string) (string, error) {
	// Find the image id by listing all images and matching on name.
	var imageId string
	images, err := self.getBlockDeviceTemplateGroups()
	if err != nil {
		return "", err
	}

	for _, val := range images {
		image := val.(map[string]interface{})
		if image["name"] == imageName && image["globalIdentifier"] != nil {
			imageId = image["globalIdentifier"].(string)
			break
		}
	}

	if imageId == "" {
		err = fmt.Errorf("No image found with name '%s'.", imageName)
		return "", err
	}

	return imageId, nil;
}


func (self SoftlayerClient) captureStandardImage(instanceId string, imageName string, imageDescription string, blockDeviceIds []int64) (map[string]interface{}, error) {
	blockDevices := make([]*BlockDevice, len(blockDeviceIds))
	for i, id := range blockDeviceIds {
		blockDevices[i] = &BlockDevice{
			Id: id,
		}
	}

	requestBody, err := self.generateRequestBody(imageName, blockDevices, imageDescription)
	if err != nil {
		return nil, err
	}

	data, err := self.doHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/createArchiveTransaction.json", instanceId), "POST", requestBody)
	if err != nil {
		return nil, err
	}

	return data[0].(map[string]interface{}), err
}

func (self SoftlayerClient) captureImage(instanceId string, imageName string, imageDescription string) (map[string]interface{}, error) {
	imageRequest := &InstanceImage{
		Descption: imageDescription,
		Name:      imageName,
		Summary:   imageDescription,
	}

	requestBody, err := self.generateRequestBody(imageRequest)
	if err != nil {
		return nil, err
	}

	data, err := self.doHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/captureImage.json", instanceId), "POST", requestBody)
	if err != nil {
		return nil, err
	}

	return data[0].(map[string]interface{}), err
}

func (self SoftlayerClient) destroyImage(imageId string) error {
	response, err := self.doRawHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s.json", imageId), "DELETE", new(bytes.Buffer))

	log.Printf("Deleted an image with id (%s), response: %s", imageId, response)
	if res := string(response[:]); res != "true" {
		return errors.New(fmt.Sprintf("Failed to destroy and image wit id '%s', got '%s' as response from the API.", imageId, res))
	}

	return err
}

func (self SoftlayerClient) isInstanceReady(instanceId string) (bool, error) {
	powerData, err := self.doHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/getPowerState.json", instanceId), "GET", nil)
	if err != nil {
		return false, nil
	}
	isPowerOn := powerData[0].(map[string]interface{})["keyName"].(string) == "RUNNING"

	transactionData, err := self.doHttpRequest(fmt.Sprintf("SoftLayer_Virtual_Guest/%s/getActiveTransaction.json", instanceId), "GET", nil)
	if err != nil {
		return false, nil
	}
	noTransactions := transactionData[0] == nil

	return isPowerOn && noTransactions, err
}

func (self SoftlayerClient) waitForInstanceReady(instanceId string, timeout time.Duration) error {
	done := make(chan struct{})
	defer close(done)
	result := make(chan error, 1)

	go func() {
		attempts := 0
		for {
			attempts += 1

			log.Printf("Checking instance status... (attempt: %d)", attempts)
			isReady, err := self.isInstanceReady(instanceId)
			if err != nil {
				result <- err
				return
			}

			if isReady {
				result <- nil
				return
			}

			// Wait 3 seconds in between
			time.Sleep(3 * time.Second)

			// Verify we shouldn't exit
			select {
			case <-done:
				// We finished, so just exit the goroutine
				return
			default:
				// Keep going
			}
		}
	}()

	log.Printf("Waiting for up to %d seconds for instance to become ready", timeout/time.Second)
	select {
	case err := <-result:
		return err
	case <-time.After(timeout):
		err := fmt.Errorf("Timeout while waiting to for the instance to become ready")
		return err
	}
}
