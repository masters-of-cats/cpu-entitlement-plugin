.PHONY: test

build:
	go build

test:
	ginkgo -r --race

install:
	cf uninstall-plugin CpuEntitlementPlugin || true
	cf install-plugin ./cpu-entitlement-plugin -f
