package softlayer

import (
	"testing"
)

func TestClient_FindNonSwapBlockDeviceIds(t *testing.T) {
	client := SoftlayerClient{}

	result := client.findNonSwapBlockDeviceIds(
		[]interface{} {
			map[string]interface{} {
				"device": "0",
				"id": 11.,
				"diskImage": map[string]interface{} {
					"name": "root-device",
					"id": 12.,
				},
			},
			map[string]interface{} {
				"device": "1",
				"id": 21.,
				"diskImage": map[string]interface{} {
					"name": "SWAP-device",
					"id": 22.,
				},
			},
		})
	if len(result) != 1 {
		t.Fatalf("Expected only one device but got '%v'", result)
	}
	if result[0] != 11 {
		t.Fatalf("Expected device id 11 but got %d", result[0])
	}



	result = client.findNonSwapBlockDeviceIds(
		[]interface{} {
			map[string]interface{} {
				"device": "0",
				"id": 11.,
				"diskImage": map[string]interface{} {
					"name": "first-SWAP-device",
					"id": 12.,
				},
			},
			map[string]interface{} {
				"device": "1",
				"id": 21.,
				"diskImage": map[string]interface{} {
					"name": "SWAP-device",
					"id": 22.,
				},
			},
		})
	if len(result) != 0 {
		t.Fatalf("Expected no devices but got '%v'", result)
	}



	result = client.findNonSwapBlockDeviceIds(
		[]interface{} {
			map[string]interface{} {
				"device": "0",
				"id": 11.,
				"diskImage": map[string]interface{} {
					"name": "first-device",
					"id": 12.,
				},
			},
			map[string]interface{} {
				"device": "1",
				"id": 21.,
				"diskImage": map[string]interface{} {
					"name": "second-device",
					"id": 22.,
				},
			},
		})
	if len(result) != 2 {
		t.Fatalf("Expected two devices but got '%v'", result)
	}
	if result[0] != 11 || result[1] != 21 {
		t.Fatalf("Expected devices 11 and 21 but got %v", result)
	}
}

