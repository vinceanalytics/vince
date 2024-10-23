RELEASE_NAME := vince-vultr-server
VINCE_VERSION ?= $(shell git describe --abbrev=0 --tags)

.PHONY: $(MAKECMDGOALS)

release-vince-vultr-server:
	cp ./vince/etc/update-motd.d/99-one-click.tpl ./vince/etc/update-motd.d/99-one-click
	sed -i -e "s/VINCE_VERSION/${VINCE_VERSION}/g" ./vince/etc/update-motd.d/99-one-click
	packer init vince/vince.pkr.hcl
	packer build vince/vince.pkr.hcl