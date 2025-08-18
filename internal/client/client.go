package client

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/rtamalin/rmt-client-testing/internal/choice"
	"github.com/rtamalin/rmt-client-testing/internal/profile"
)

type ClientId uint64
type ClientType uint

const (
	CLIENT_TINY ClientType = iota
	CLIENT_SMALL
	CLIENT_MEDIUM
	CLIENT_LARGE
	CLIENT_METAL
	numClientTypes
)

var clientTypes = []string{
	"tiny",
	"small",
	"medium",
	"large",
	"metal",
	"UNKNOWN",
}

func (ct ClientType) String() string {
	if ct < numClientTypes {
		return clientTypes[ct]
	}
	return clientTypes[numClientTypes]
}

type Client struct {
	Id      ClientId
	Name    string
	UUID    string
	Type    ClientType
	NumDisk int
	NumGPU  int
	NumNet  int
	PciData *profile.ProfileInfo
	ModList *profile.ProfileInfo
}

type ClientHwInfo struct {
	Arch    string
	Cpus    int
	Memory  int
	Sockets int
}

var HwInfo = []*ClientHwInfo{
	{ // CLIENT_TINY
		Arch:    "x86_64",
		Cpus:    2,
		Memory:  512,
		Sockets: 1,
	},
	{ // CLIENT_SMALL
		Arch:    "x86_64",
		Cpus:    2,
		Memory:  1024,
		Sockets: 1,
	},
	{ // CLIENT_MEDIUM
		Arch:    "x86_64",
		Cpus:    2,
		Memory:  8 * 1024,
		Sockets: 1,
	},
	{ // CLIENT_LARGE
		Arch:    "x86_64",
		Cpus:    4,
		Memory:  16 * 1024,
		Sockets: 1,
	},
	{ // CLIENT_METAL
		Arch:    "x86_64",
		Cpus:    96,
		Memory:  384 * 1024,
		Sockets: 4,
	},
}

type newClientFunc func(ClientId) *Client

var ClientChoices = [numClientTypes]choice.Choice{
	{
		Weight: 20,
		Value:  newClientFunc(TinyClient),
	},
	{
		Weight: 20,
		Value:  newClientFunc(SmallClient),
	},
	{
		Weight: 20,
		Value:  newClientFunc(MediumClient),
	},
	{
		Weight: 20,
		Value:  newClientFunc(LargeClient),
	},
	{
		Weight: 20,
		Value:  newClientFunc(MetalClient),
	},
}

func NewClient(id ClientId) *Client {
	clientFunc := choice.Choose(ClientChoices[:]).(newClientFunc)

	return clientFunc(id)
}

func (c *Client) Init(cliType ClientType, id ClientId, numDisk, numGPU, numNet int) {
	c.Id = id
	c.Type = cliType
	c.NumDisk = numDisk
	c.NumGPU = numGPU
	c.NumNet = numNet

	c.Name = fmt.Sprintf("%s-%d", cliType, id)
	c.UUID = uuid.New().String()
}

func (c *Client) Uname() string {
	return fmt.Sprintf(
		"Simulated %s client %s with %d Disks, %d GPUs, %d Nets",
		c.Type,
		c.UUID,
		c.NumDisk,
		c.NumGPU,
		c.NumNet,
	)
}

func (c *Client) SystemInfo() string {
	sysInfo := make(map[string]any)
	hwInfo := HwInfo[c.Type]

	sysInfo["arch"] = hwInfo.Arch
	sysInfo["cloud_provider"] = "amazon"
	sysInfo["cpus"] = hwInfo.Cpus
	sysInfo["hostname"] = c.Id
	sysInfo["hypervisor"] = "amazon"
	sysInfo["mem_total"] = hwInfo.Memory
	sysInfo["mod_list"] = c.ModList
	sysInfo["pci_data"] = c.PciData
	sysInfo["sockets"] = hwInfo.Sockets
	sysInfo["uname"] = c.Uname()
	sysInfo["uuid"] = c.UUID

	siBytes, err := json.Marshal(sysInfo)
	if err != nil {
		log.Fatalf(
			"Failed to generate system info JSON for %s client %s: %s",
			c.Type,
			c.UUID,
			err.Error(),
		)
	}

	return string(siBytes)
}

func (c *Client) setupPciData(header []string, pciBus, pciSlot int) {
	// allocate pciData with capacity to hold header plus sufficent
	// lines for the added entries
	pciData := make([]string, 0, len(header)+c.NumDisk+c.NumGPU+c.NumNet)

	// copy header elements
	pciData = append(pciData, header...)

	// add disk devices as next slot in same bus
	for i := 0; i < c.NumDisk; i++ {
		pciEntry := fmt.Sprintf(
			"%02x:%02x.0 Non-Volatile memory controller: Amazon.com, Inc. NVMe EBS Controller",
			pciBus,
			pciSlot,
		)
		pciSlot++ // increment the slot
		pciData = append(pciData, pciEntry)
	}

	// add gpu devices as next slot in same bus
	for i := 0; i < c.NumGPU; i++ {
		pciEntry := fmt.Sprintf(
			"%02x:%02x.0 3D controller: NVIDIA Corporation TU104GL [Tesla T4] (rev a1)",
			pciBus,
			pciSlot,
		)
		pciSlot++ // increment the slot
		pciData = append(pciData, pciEntry)
	}

	// add network devices as next slot in same bus
	for i := 0; i < c.NumNet; i++ {
		pciEntry := fmt.Sprintf(
			"%02x:%02x.0 Ethernet controller: Amazon.com, Inc. Elastic Network Adapter (ENA)",
			pciBus,
			pciSlot,
		)
		pciSlot++ // increment the slot
		pciData = append(pciData, pciEntry)
	}

	pciData = append(pciData, "" /* blank last line creates trailing newline */)

	c.PciData = profile.NewProfileInfo(strings.Join(pciData, "\n"))
}

func (c *Client) setupModList(modList []string) {
	ml := make([]string, len(modList))
	copy(ml, modList)
	c.ModList = profile.NewProfileInfo(ml)
}
