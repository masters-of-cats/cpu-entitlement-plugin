.PHONY: test

build:
	go build

test:
	ginkgo -r --race

install:
	cf uninstall-plugin CPUEntitlementPlugin2 || true
	cf install-plugin ./cpu-entitlement-plugin -f
