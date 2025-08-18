package client

import (
	"github.com/rtamalin/rmt-client-testing/internal/choice"
)

var smallPciDataHeader = []string{
	"00:00.0 Host bridge: Intel Corporation 440FX - 82441FX PMC [Natoma]",
	"00:01.0 ISA bridge: Intel Corporation 82371SB PIIX3 ISA [Natoma/Triton II]",
	"00:03.0 VGA compatible controller: Amazon.com, Inc. Device 1111",
}

const (
	smallPciBus  = 0
	smallPciSlot = 4
)

var smallModList = []string{
	"aesni_intel",
	"af_packet",
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
	"parport",
	"parport_pc",
	"pcspkr",
	"ppdev",
	"rfkill",
	"serio_raw",
	"sha1_ssse3",
	"sha256_ssse3",
	"sha512_ssse3",
	"sunrpc",
	"t10_pi",
	"vfat",
	"xfs",
	"x_tables",
}

// weighted choice of number of disks for a small client
var smallDiskChoices = []choice.Choice{
	{
		Weight: 60,
		Value:  1,
	},
	{
		Weight: 30,
		Value:  2,
	},
	{
		Weight: 10,
		Value:  3,
	},
}

func SmallClient(id ClientId) *Client {
	c := new(Client)

	numDisk := choice.Choose(smallDiskChoices).(int)
	numGPU := 0
	numNet := 1

	c.Init(CLIENT_SMALL, id, numDisk, numGPU, numNet)

	c.setupPciData(smallPciDataHeader, smallPciBus, smallPciSlot)
	c.setupModList(smallModList)

	return c
}
