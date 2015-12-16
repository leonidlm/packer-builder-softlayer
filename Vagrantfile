# -*- mode: ruby -*-
# vi: set ft=ruby :
# Original version of this file copied from: https://raw.githubusercontent.com/mitchellh/packer/master/Vagrantfile
#

# VM Specifications
VM_MEMORY=2048
VM_CPUS=2
VM_GUI=false

GOROOT = '/opt/go'
GOPATH = '/opt/gopath'
PACKAGE_PATH = 'src/github.com/leonidlm/packer-builder-softlayer'

script = <<SCRIPT
SRCROOT="#{GOROOT}"

# Install Go
sudo apt-get update
sudo apt-get install -y build-essential mercurial
sudo hg clone -u release https://code.google.com/p/go $SRCROOT
cd ${SRCROOT}/src
sudo ./all.bash

# Setup the GOPATH
sudo mkdir -p #{GOPATH}
cat <<EOF >/tmp/gopath.sh
export GOPATH="#{GOPATH}"
export PATH="#{GOROOT}/bin:\\$GOPATH/bin:\\$PATH"
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

# Build packer-builder-softlayer
cd $GOPATH/#{PACKAGE_PATH}
go build
go test ./...
go install

# Make sure the gopath is usable by vagrant
sudo chown -R -f vagrant:vagrant $SRCROOT
sudo chown -R -f vagrant:vagrant #{GOPATH} 

echo "Ready for development. Begin with cd $GOPATH/#{PACKAGE_PATH}"

SCRIPT

Vagrant.configure(2) do |config|
  config.vm.box = "bento/ubuntu-12.04"

  config.vm.synced_folder '.', "#{GOPATH}/#{PACKAGE_PATH}", id: 'src' 

  config.vm.provision 'shell', inline: script

  config.vm.provider :virtualbox do |vb|
    vb.memory = VM_MEMORY
    vb.cpus = VM_CPUS
    vb.gui = VM_GUI
  end

  config.vm.provider 'parallels' do |parallels, override|
    override.vm.box = 'parallels/ubuntu-12.04'
    override.vm.synced_folder '.', "#{GOPATH}/#{PACKAGE_PATH}", mount_options: ['rw', 'nosuid', 'nodev', 'sync', 'noatime', 'share'], id: 'src'
    parallels.name = 'Packer SoftLayer Development Box'
    parallels.optimize_power_consumption = false
    parallels.memory = VM_MEMORY
    parallels.cpus = VM_CPUS
  end
end

