# ComputeProviderByGo
This project is the full platform client of [BountyCloud](https://www.bountycloud.net)

Security needs to delete AK, SK and other key information

## How to run? 

### Step1 
Install docker and run service

>ubuntu: https://www.runoob.com/docker/ubuntu-docker-install.html
> 
>win: https://www.runoob.com/docker/windows-docker-install.html
> 
>mac: https://www.runoob.com/docker/macos-docker-install.html

### Step2

> Download the latest version of the client in the release

### Step3 

#### mac&linux
In terminal run:

>go_mac_provider_amd64 --username=xxxx --password=xxxx --cpu=1 --ram=1
> 
>go_linux_provider_amd64 --username=xxxx --password=xxxx --cpu=1 --ram=1

#### win
In cmd run:

> start go_linux_provider_amd64 --username=xxxx --password=xxxx --cpu=1 --ram=1

## ISSUE

### win port bind fail

In cmd run:

> net stop hns
> 
> net start hns 
> 
> net stop winnat
> 
> start docket 
> 
> net start winnat

### can't get the right public IP?
> Check your pc firewall and make sure the icmp protocol is turned on