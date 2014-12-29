package softlayer

import (
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{} {
		"username":     "test",
		"api_key":      "testkey",
		"image_name":   "testimage",
		"base_os_code": "TEST_OS_CODE",
	}
}

func TestPrepare_ImageType(t *testing.T) {
	var b Builder

	c := testConfig()

	// Default image_type
	if _, err := b.Prepare(c); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if b.config.ImageType != "flex" {
		t.Fatalf("Expected default image_type 'flex' but got '%s'", b.config.ImageType)
	}

	// Verify standard images are supported
	c["image_type"] = "standard"
	if _, err := b.Prepare(c); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if b.config.ImageType != "standard" {
		t.Fatalf("Expected image_type 'standard' but got '%s'", b.config.ImageType)
	}

	// Unknown image_type
	c["image_type"] = "super"
	if _, err := b.Prepare(c); err == nil {
		t.Fatal("Expected an error")
	}
}

