package zram

import (
	"strconv"
)

type ZramConfigurations struct {
	Name  string
	Disks []ZramDiskConfiguration
}

type ZramDiskConfiguration struct {
	ID                    string
	Host_memory_limit     string
	Zram_fraction         float64
	Max_zram_size         int
	Compression_algorithm string
	Fs_type               string
	Swap_priority         int
	Mount_point           string
}

func (zrd ZramDiskConfiguration) ToWrite() []string {
	if zrd.Fs_type == SWAP {
		strings := []string{
			"[" + zrd.ID + "]\n",
			"host-memory-limit = " + zrd.Host_memory_limit + "\n",
			"zram-fraction = " + strconv.FormatFloat(zrd.Zram_fraction, 'f', -1, 64) + "\n",
			"max-zram-size = " + strconv.Itoa(zrd.Max_zram_size) + "\n",
			"compression-algorithm = " + zrd.Compression_algorithm + "\n",
			"fs-type = " + zrd.Fs_type + "\n",
			"swap-priority = " + strconv.Itoa(zrd.Swap_priority) + "\n",
		}
		return strings
	}
	/*if zrd.Mount_point == "" {
		zrd.Mount_point = TMP
	}*/
	strings := []string{
		"[" + zrd.ID + "]\n",
		"host-memory-limit = " + zrd.Host_memory_limit + "\n",
		"zram-fraction = " + strconv.FormatFloat(zrd.Zram_fraction, 'f', -1, 64) + "\n",
		"max-zram-size = " + strconv.Itoa(zrd.Max_zram_size) + "\n",
		"compression-algorithm = " + zrd.Compression_algorithm + "\n",
		"fs-type = " + zrd.Fs_type + "\n",
		"mount-point = " + zrd.Mount_point + "\n",
	}
	return strings
}

func (zrc ZramConfigurations) ToWrite() []string {
	var strings []string
	for _, zrd := range zrc.Disks {
		strings = append(strings, "#start "+zrd.ID+" zramdisk configuration \n\n")
		strings = append(strings, zrd.ToWrite()...)
		strings = append(strings, "\n#end "+zrd.ID+" zramdisk configuration \n\n")
	}
	return strings
}
