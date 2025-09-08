package client

import "github.com/rtamalin/rmt-client-testing/internal/choice"

var tinyPciDataHeader = []string{
	"00:00.0 Host bridge: Intel Corporation 440FX - 82441FX PMC [Natoma] (rev 02)",
	"00:01.0 ISA bridge: Intel Corporation 82371SB PIIX3 ISA [Natoma/Triton II]",
	"00:02.0 VGA compatible controller: Cirrus Logic GD 5446",
	"00:03.0 Unassigned class [ff80]: XenSource, Inc. Xen Platform Device (rev 01)",
}

const (
	tinyPciBus  = 0
	tinyPciSlot = 4
)

var tinyModData = []string{
	"aesni_intel",
	"af_packet",
	"ahci",
	"ata_generic",
	"ata_piix",
	"blake2b_generic",
	"btrfs",
	"button",
	"cirrus",
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
	"libahci",
	"libata",
	"libcrc32c",
	"nls_cp437",
	"nls_iso8859_1",
	"pcspkr",
	"raid6_pq",
	"rfkill",
	"scsi_mod",
	"sd_mod",
	"serio_raw",
	"sg",
	"sha1_ssse3",
	"sha256_ssse3",
	"sha512_ssse3",
	"sunrpc",
	"t10_pi",
	"vfat",
	"xen_blkfront",
	"xen_netfront",
	"xfs",
	"xor",
	"x_tables",
}

// weighted choice of number of disks for a small client
var tinyDiskChoices = []choice.Choice{
	{
		Weight: 50,
		Value:  1,
	},
	{
		Weight: 25,
		Value:  2,
	},
	{
		Weight: 10,
		Value:  3,
	},
	{
		Weight: 10,
		Value:  4,
	},
	{
		Weight: 5,
		Value:  5,
	},
}

func TinyClient(id ClientId) *Client {
	c := new(Client)

	numDisk := choice.Choose(tinyDiskChoices).(int)
	numGPU := 0
	numNet := 0

	c.Init(CLIENT_TINY, id, numDisk, numGPU, numNet)

	c.setupPciData(tinyPciDataHeader, tinyPciBus, tinyPciSlot)
	c.setupModData(tinyModData)

	return c
}
