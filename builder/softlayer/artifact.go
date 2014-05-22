package softlayer

import (
  "fmt"
  "errors"
)

// Artifact represents a Softlayer image as the result of a Packer build.
type Artifact struct {
  imageName string
}

// BuilderId returns the builder Id.
func (*Artifact) BuilderId() string {
  return BuilderId
}

// Destroy destroys the Softlayer image represented by the artifact.
func (a *Artifact) Destroy() error {
  fmt.Println("Destroying image: %s", a.imageName)
  e := errors.New("error happened now!")
  return e
}

// Files returns the files represented by the artifact.
func (*Artifact) Files() []string {
  return nil
}

// Id returns the Softlayer image name.
func (a *Artifact) Id() string {
  return a.imageName
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
  return fmt.Sprintf("A disk image was created: %v", a.imageName)
}