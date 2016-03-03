# Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
#                   The Finc Authors, http://finc.info
#                   Martin Czygan, <martin.czygan@uni-leipzig.de>
#
# This file is part of some open source application.
# 
# Some open source application is free software: you can redistribute
# it and/or modify it under the terms of the GNU General Public
# License as published by the Free Software Foundation, either
# version 3 of the License, or (at your option) any later version.
# 
# Some open source application is distributed in the hope that it will
# be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
# of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
# 
# You should have received a copy of the GNU General Public License
# along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
# 
# @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
# 

# -*- mode: ruby -*-
# vi: set ft=ruby :

$script = <<SCRIPT
#                  _
#   ___ _ __   ___| |
#  / _ \ '_ \ / _ \ |
# |  __/ |_) |  __/ |
#  \___| .__/ \___|_|
#      |_|
#

sudo rpm -ivh http://dl.fedoraproject.org/pub/epel/6/x86_64/epel-release-6-8.noarch.rpm
sudo yum -y update

#    ___       _   _
#   / _ \_   _| |_| |__   ___  _ __
#  / /_)/ | | | __| '_ \ / _ \| '_ \
# / ___/| |_| | |_| | | | (_) | | | |
# \/     \__, |\__|_| |_|\___/|_| |_|
#        |___/
#

sudo yum groupinstall -y 'development tools'
sudo yum install -y zlib-dev openssl-devel sqlite-devel bzip2-devel xz-libs
cd /tmp
wget https://www.python.org/ftp/python/2.7.9/Python-2.7.9.tar.xz
tar xf Python-2.7.9.tar.xz
cd Python-2.7.9
./configure --prefix=/usr/local && make && sudo make altinstall
cd /usr/local/bin
sudo ln -s python2.7 python
sudo ln -s python2.7-config python-config
cd /tmp
wget https://bitbucket.org/pypa/setuptools/raw/bootstrap/ez_setup.py
sudo /usr/local/bin/python2.7 ez_setup.py
sudo /usr/local/bin/easy_install-2.7 pip
[ ! -e "/etc/profile.d/extrapath.sh" ] && sudo sh -c 'echo "export PATH=\"/usr/local/bin:/usr/local/sbin:$PATH\"" > /etc/profile.d/extrapath.sh'

#      _     _    _
#  ___(_)___| | _(_)_ __
# / __| / __| |/ / | '_ \
# \__ \ \__ \   <| | | | |
# |___/_|___/_|\_\_|_| |_|
#

sudo yum install -y libxml2 libxslt libxml2-devel libxslt-devel
sudo -i pip install --upgrade --no-cache-dir siskin
sudo mkdir -p /etc/bash_completion.d/
sudo wget -O /etc/bash_completion.d/siskin_completion.sh https://raw.githubusercontent.com/miku/siskin/master/contrib/siskin_completion.sh
sudo chmod +x /etc/bash_completion.d/siskin_completion.sh

#    __      _
#   /__\_  _| |_ _ __ __ _
#  /_\ \ \/ / __| '__/ _` |
# //__  >  <| |_| | | (_| |
# \__/ /_/\_\\__|_|  \__,_|
#
# extra tools required by siskin or just useful

sudo yum install -y jq xmlstarlet lftp vim tmux bash-completion tree

SCRIPT

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  config.ssh.insert_key = false
  config.vm.box = "puphpet/centos65-x64"
  config.vm.provision "shell", inline: $script
end
