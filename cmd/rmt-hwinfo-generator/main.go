package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/rtamalin/rmt-client-testing/internal/client"
	"github.com/rtamalin/rmt-client-testing/internal/clientstore"
	"github.com/rtamalin/rmt-client-testing/internal/profile"
)

const (
	CREATE = true
)

type Options struct {
	NumClients int64
	DataStore  string
}

var option_defaults = Options{
	DataStore:  "ClientDataStore",
	NumClients: 1000,
}

var options Options

type ProfileInfoStatBlock struct {
	Size     int `json:"size"`
	JsonSize int `json:"jsonSize"`
	Count    int `json:"count"`
}

type ProfileInfoStats *ProfileInfoStatBlock

func NewProfileKeyStats(pInfo *profile.ProfileInfo) ProfileInfoStats {
	p := new(ProfileInfoStatBlock)

	// init to 0, will be incremented later
	p.Count = 0

	// determine the storage size for storing the profile's data, which
	// should be JSON encoded string...
	switch v := pInfo.Data.(type) {
	case []string:
		p.Size = len(strings.Join(v, "\n"))
	case string:
		p.Size = len(v)
	default:
		log.Fatalf("ERROR: Unsupported profile data type %T for %v", v, v)
	}

	infoMap := map[string]any{
		"digest": pInfo.Digest,
		"data":   pInfo.Data,
	}

	withData, _ := json.Marshal(infoMap)

	delete(infoMap, "data")
	withoutData, _ := json.Marshal(infoMap)

	p.JsonSize = len(withData) - len(withoutData)

	return p
}

type HwInfoStats struct {
	ProfileStats       map[string]map[string]ProfileInfoStats `json:"profileStats"`
	NumProfileTypes    int                                    `json:"numProfileTypes"`
	NumUniqueProfiles  int                                    `json:"numUniqueProfiles"`
	ProfileStorageSize int                                    `json:"profileStorageSize"`
	HwInfoSavings      int                                    `json:"hwInfoSavings"`
	DbNetSavings       int                                    `json:"dbNetSavings"`
}

func NewHwInfoStats() *HwInfoStats {
	h := new(HwInfoStats)
	h.Init()
	return h
}

func (h *HwInfoStats) Init() {
	h.ProfileStats = make(map[string]map[string]ProfileInfoStats)
}

func (h *HwInfoStats) Add(profileName string, pInfo *profile.ProfileInfo) {
	// add an entry for the profile if not seen before
	if _, exists := h.ProfileStats[profileName]; !exists {
		h.ProfileStats[profileName] = make(map[string]ProfileInfoStats)
	}
	if _, exists := h.ProfileStats[profileName][pInfo.Digest]; !exists {
		h.ProfileStats[profileName][pInfo.Digest] = NewProfileKeyStats(pInfo)
	}

	h.ProfileStats[profileName][pInfo.Digest].Count++
}

func (h *HwInfoStats) Finalize() {
	h.NumProfileTypes = len(h.ProfileStats)
	h.NumUniqueProfiles = 0
	h.ProfileStorageSize = 0
	h.HwInfoSavings = 0
	for typeName, typeMap := range h.ProfileStats {
		h.NumUniqueProfiles += len(typeMap)
		for pId, pStats := range typeMap {
			// add the approx size to store the data profile entry itself
			h.ProfileStorageSize += (8 /* approx size of primary key id field */ +
				len(typeName) +
				len(pId) +
				pStats.Size +
				18 /* approx size of three timestamp fields, createdAt, updatedAt, lastSeen */)

			// calculate the savings gained by not storing the profile data
			// in the hwinfo JSON blob.
			h.HwInfoSavings += (pStats.Count - 1) * pStats.JsonSize
		}
	}
	h.DbNetSavings = h.HwInfoSavings - h.ProfileStorageSize
}

func (h *HwInfoStats) Write(dsDir string) (err error) {
	filePath := filepath.Join(dsDir, "HwInfoStats.json")

	hwisData, err := json.Marshal(h)
	if err != nil {
		err = fmt.Errorf(
			"failed to json.Marshall() HwInfoStats: %w",
			err,
		)
		return
	}

	err = os.WriteFile(filePath, hwisData, 0o644)
	if err != nil {
		err = fmt.Errorf(
			"failed to write HwInfoStats to %q: %w",
			filePath,
			err,
		)
		return
	}

	return
}

func main() {
	options = option_defaults
	flag.StringVar(&options.DataStore, "datastore", option_defaults.DataStore, "Location of `datastore` to store simulated clients")
	flag.Int64Var(&options.NumClients, "clients", option_defaults.NumClients, "The number of `clients` to simulate")
	flag.Parse()

	if (options.NumClients >= math.MaxUint32) || (options.NumClients < 0) {
		log.Fatal(
			"ERROR: The number of clients must be a positive value between 0 and MaxUint32\n",
		)
	}

	log.Printf("Initialising %q as datastore\n", options.DataStore)
	dataStore := clientstore.New(options.DataStore)

	log.Printf("Simulating %v clients\n", options.NumClients)

	hwInfoStats := NewHwInfoStats()
	for i := int64(0); i < options.NumClients; i++ {
		c := client.NewClient(client.ClientId(i))
		sysInfo := c.SystemInfo()

		fileId := clientstore.FileId(i)
		fileType := clientstore.SYS_INFO_TYPE
		err := dataStore.WriteFile(fileId, fileType, []byte(sysInfo), 0o644)
		if err != nil {
			log.Fatalf(
				"Failed to write client %v to %q: %s",
				i,
				fileId.Path(fileType),
				err.Error(),
			)
		}

		hwInfoStats.Add(client.MOD_DATA_PROFILE, c.ModData)
		hwInfoStats.Add(client.PCI_DATA_PROFILE, c.PciData)
	}

	hwInfoStats.Finalize()
	if err := hwInfoStats.Write(options.DataStore); err != nil {
		log.Fatalf(
			"Failed to record client hardware info stats for %d generated clients under %q",
			options.NumClients,
			options.DataStore,
		)
	}

	log.Printf(
		"Generated hardware info for %d clients under %q",
		options.NumClients,
		options.DataStore,
	)
}
