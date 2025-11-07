package main

import (
	"github.com/SUSE/connect-ng/pkg/registration"
)

func prepareExtraData(sysInfo SysInfo, cliOpts *CliOpts) registration.ExtraData {
	extraData := registration.ExtraData{
		"instance_data": cliOpts.instData,
	}

	if cliOpts.NoDataProfiles {
		return extraData
	}

	// add system profiles to extraData.dataProfiles, removing them from sysInfo
	systemProfiles := map[string]any{}
	spNames := []string{
		"pci_data",
		"mod_list",
	}
	for _, spName := range spNames {
		// skip if spName entry not in sysInfo
		if _, ok := sysInfo[spName]; !ok {
			continue
		}

		systemProfiles[spName] = sysInfo[spName]
		delete(sysInfo, spName)
	}
	extraData["system_profiles"] = systemProfiles

	return extraData
}
