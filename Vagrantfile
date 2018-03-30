# -*- mode: ruby -*-
# vi: set ft=ruby :
# Original version of this file copied from: https://raw.githubusercontent.com/hashicorp/packer/master/Vagrantfile
#

# VM Specifications
VM_MEMORY=2048
VM_CPUS=2
VM_GUI=false

GOROOT = '/opt/go'
GOPATH = '/opt/gopath'
PACKAGE_PATH = 'src/github.com/watson-platform/packer-builder-softlayer'
GO_VERSION = '1.9.2'

script = <<SCRIPT
set -x

SRCROOT="#{GOROOT}"

# Install Go

download="https://storage.googleapis.com/golang/go#{GO_VERSION}.linux-amd64.tar.gz"

wget -q -O /tmp/go.tar.gz ${download}

tar -C /tmp -xf /tmp/go.tar.gz
sudo mv /tmp/go /usr/local
sudo chown -R root:root /usr/local/go

# Ensure that the GOPATH tree is owned by vagrant:vagrant
mkdir -p #{GOPATH}
sudo chown -R vagrant:vagrant #{GOPATH}
sudo chown -R -f vagrant:vagrant $SRCROOT

# Ensure Go is on PATH
if [ ! -e /usr/bin/go ] ; then
  ln -s /usr/local/go/bin/go /usr/bin/go
fi
if [ ! -e /usr/bin/gofmt ] ; then
  ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt
fi


# Ensure new sessions know about GOPATH
if [ ! -f /etc/profile.d/gopath.sh ] ; then
  cat <<EOT > /etc/profile.d/gopath.sh
export GOPATH="#{GOPATH}"
export PATH="#{GOPATH}/bin:\$PATH"
EOT
  chmod 755 /etc/profile.d/gopath.sh
fi

source /etc/profile.d/gopath.sh


# Install some other stuff we need
sudo apt-get update && sudo apt-get install -y curl git-core zip build-essential

# Download and build Packer
go get -u github.com/mitchellh/gox
gox -build-toolchain
go get -d -u github.com/hashicorp/packer
cd $GOPATH/src/github.com/hashicorp/packer
make

# Build packer-builder-softlayer
cd $GOPATH/#{PACKAGE_PATH}
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
dep ensure
go build
go test ./...
go install

echo "Ready for development. Begin with cd $GOPATH/#{PACKAGE_PATH}"

SCRIPT

Vagrant.configure(2) do |config|
  config.vm.box = "ubuntu/trusty64"

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

