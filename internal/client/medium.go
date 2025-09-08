package client

import (
	"github.com/rtamalin/rmt-client-testing/internal/choice"
)

var mediumPciDataHeader = []string{
	"00:00.0 Host bridge: Intel Corporation 440FX - 82441FX PMC [Natoma]",
	"00:01.0 ISA bridge: Intel Corporation 82371SB PIIX3 ISA [Natoma/Triton II]",
	"00:03.0 VGA compatible controller: Amazon.com, Inc. Device 1111",
}

const (
	mediumPciBus  = 0 // PCI Bus to add new entries to
	mediumPciSlot = 4 // Starting PCI Slot for new entries
)

var mediumModData = []string{
	"aesni_intel",
	"af_packet",
	"blake2b_generic",
	"btrfs",
	"button",
	"configfs",
	"crc32c_intel",
	"crc32_pclmul",
	"crc64",
	"crc64_rocksoft",
	"crc64_rocksoft_generic",
	"cryptd",
	"crypto_simd",
	"dmi_sysfs",
	"dm_log",
	"dm_mirror",
	"dm_mod",
	"dm_region_hash",
	"efivarfs",
	"ena",
	"fat",
	"fuse",
	"ghash_clmulni_intel",
	"i2c_piix4",
	"intel_rapl_common",
	"intel_rapl_msr",
	"intel_uncore_frequency_common",
	"ip_tables",
	"iscsi_boot_sysfs",
	"iscsi_ibft",
	"libcrc32c",
	"libnvdimm",
	"nfit",
	"nls_cp437",
	"nls_iso8859_1",
	"nvme",
	"nvme_auth",
	"nvme_core",
	"nvme_keyring",
	"parport",
	"parport_pc",
	"pcspkr",
	"ppdev",
	"raid6_pq",
	"rfkill",
	"serio_raw",
	"sha1_ssse3",
	"sha256_ssse3",
	"sha512_ssse3",
	"sunrpc",
	"t10_pi",
	"vfat",
	"xfs",
	"xor",
	"x_tables",
}

// weighted choice of number of disks for a medium client
var mediumDiskChoices = []choice.Choice{
	{
		Weight: 45,
		Value:  1,
	},
	{
		Weight: 35,
		Value:  2,
	},
	{
		Weight: 20,
		Value:  3,
	},
}

var mediumGPUChoices = []choice.Choice{
	{
		Weight: 85,
		Value:  0,
	},
	{
		Weight: 10,
		Value:  1,
	},
	{
		Weight: 5,
		Value:  2,
	},
}

func MediumClient(id ClientId) *Client {
	c := new(Client)

	numDisk := choice.Choose(mediumDiskChoices).(int)
	numGPU := choice.Choose(mediumGPUChoices).(int)
	numNet := 1

	c.Init(CLIENT_MEDIUM, id, numDisk, numGPU, numNet)

	c.setupPciData(mediumPciDataHeader, mediumPciBus, mediumPciSlot)
	c.setupModData(mediumModData)

	return c
}
