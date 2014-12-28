# -*- mode: ruby -*-
# vi: set ft=ruby :
# Original version of this file copied from: https://raw.githubusercontent.com/mitchellh/packer/master/Vagrantfile
#

$script = <<SCRIPT
SRCROOT="/opt/go"

# Install Go
sudo apt-get update
sudo apt-get install -y build-essential mercurial
sudo hg clone -u release https://code.google.com/p/go $SRCROOT
cd ${SRCROOT}/src
sudo ./all.bash

# Setup the GOPATH
sudo mkdir -p /opt/gopath
cat <<EOF >/tmp/gopath.sh
export GOPATH="/opt/gopath"
export PATH="/opt/go/bin:\\$GOPATH/bin:\\$PATH"
export SL_USERNAME=#{ENV['SL_USERNAME']}
export SL_API_KEY=#{ENV['SL_API_KEY']}
EOF
sudo mv /tmp/gopath.sh /etc/profile.d/gopath.sh
sudo chmod 0755 /etc/profile.d/gopath.sh

# Install some other stuff we need
sudo apt-get install -y curl git-core zip

# Download and build Packer
source /etc/profile.d/gopath.sh
go get -u github.com/mitchellh/gox
gox -build-toolchain
go get -d -u github.com/mitchellh/packer
cd $GOPATH/src/github.com/mitchellh/packer
make updatedeps
# These next two lines cause duplicate files in $GOPATH/bin
make
make dev

# Download and build packer-builder-softlayer
cd $GOPATH
# This builds packer-builder-softlayer and puts it in $GOPATH/bin
go get -d -u github.com/leonidlm/packer-builder-softlayer
cd $GOPATH/src/github.com/leonidlm/packer-builder-softlayer
#gox -os="linux windows"
go build
go install

# Make sure the gopath is usable by vagrant
sudo chown -R vagrant:vagrant $SRCROOT
sudo chown -R vagrant:vagrant /opt/gopath

SCRIPT

Vagrant.configure(2) do |config|
  config.vm.box = "chef/ubuntu-12.04"

  config.vm.provision "shell", inline: $script

  config.vm.synced_folder ".", "/vagrant", disabled: true

  ["vmware_fusion", "vmware_workstation"].each do |p|
    config.vm.provider "p" do |v|
      v.vmx["memsize"] = "2048"
      v.vmx["numvcpus"] = "2"
      v.vmx["cpuid.coresPerSocket"] = "1"
    end
  end

  config.vm.provider :virtualbox do |vb|
    vb.memory = 2048
    vb.cpus = 2
    vb.gui = false
  end
end
