package main

import (
	"github.com/SUSE/connect-ng/pkg/registration"
)

func extraDataWithDataProfiles(sysInfo SysInfo, cliOpts *CliOpts) registration.ExtraData {
	extraData := registration.ExtraData{
		"instance_data": cliOpts.instData,
	}

	// add data profiles to extraData.dataProfiles, removing them from sysInfo
	dataProfiles := map[string]any{}
	dpNames := []string{
		"pci_data",
		"mod_data",
	}
	for _, dpName := range dpNames {
		// skip if dp_name entry not in sysInfo
		if _, ok := sysInfo[dpName]; !ok {
			continue
		}

		dataProfiles[dpName] = sysInfo[dpName]
		delete(sysInfo, dpName)
	}
	extraData["data_profiles"] = dataProfiles

	return extraData
}
