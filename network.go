package main

import (
	"fmt"
	"os/exec"

	"github.com/vishvananda/netlink"
)

func setupContainerNetwork(childPid int) error {
	bridgeName := "warden0"
	bridgeIP := "172.20.0.1/24"
	hostVeth := "veth0"
	containerVeth := "veth1"

	brLink, err := netlink.LinkByName(bridgeName)
	if err != nil {
		la := netlink.NewLinkAttrs()
		la.Name = bridgeName
		br := &netlink.Bridge{LinkAttrs: la}
		if err := netlink.LinkAdd(br); err != nil {
			return fmt.Errorf("error creating bridge: %v", err)
		}
		brLink, err = netlink.LinkByName(bridgeName)
		if err != nil {
			return fmt.Errorf("error retrieving bridge link: %v", err)
		}
	}

	addr, _ := netlink.ParseAddr(bridgeIP)
	if err := netlink.AddrAdd(brLink, addr); err != nil {
		return fmt.Errorf("error adding IP address to bridge: %v", err)
	}
	if err := netlink.LinkSetUp(brLink); err != nil {
		return fmt.Errorf("error setting up bridge: %v", err)
	}

	laHost := netlink.NewLinkAttrs()
	laHost.Name = hostVeth
	veth := &netlink.Veth{
		LinkAttrs: laHost,
		PeerName:  containerVeth,
	}
	if _, err := netlink.LinkByName(hostVeth); err != nil {
		if err := netlink.LinkAdd(veth); err != nil {
			return fmt.Errorf("error creating veth pair: %v", err)
		}
	}
	vethHostLink, err := netlink.LinkByName(hostVeth)
	if err != nil {
		return fmt.Errorf("error retrieving host veth link: %v", err)
	}
	vethContainerLink, err := netlink.LinkByName(containerVeth)
	if err != nil {
		return fmt.Errorf("error retrieving container veth link: %v", err)
	}

	if err := netlink.LinkSetMaster(vethHostLink, brLink.(*netlink.Bridge)); err != nil {
		return fmt.Errorf("error setting host veth master: %v", err)
	}
	if err := netlink.LinkSetUp(vethHostLink); err != nil {
		return fmt.Errorf("error setting up host veth: %v", err)
	}

	if err := netlink.LinkSetNsPid(vethContainerLink, childPid); err != nil {
		return fmt.Errorf("error moving container veth to child namespace: %v", err)
	}
	err = setupHostNAT(bridgeIP)
	if err != nil {
		return fmt.Errorf("error setting up NAT: %v", err)
	}
	fmt.Printf("Container network setup complete with bridge %s and veth pair %s <-> %s\n", bridgeName, hostVeth, containerVeth)
	return nil

}

func setupHostNAT(bridgeSubnet string) error {
	cmd := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", bridgeSubnet, "!", "-o", "warden0", "-j", "MASQUERADE")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error setting up NAT: %v", err)
	}
	return nil
}
