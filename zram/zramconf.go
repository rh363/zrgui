package zram

import "fmt"

type ZramConfigurations struct {
	Name string
	Disk []ZramDiskConfiguration
}

type ZramDiskConfiguration struct {
	ID                    string
	Host_memory_limit     string
	Zram_fraction         float64
	Max_zram_size         string
	Compression_algorithm string
	Fs_type               string
	Swap_priority         int
	Mount_point           string
}

func NewZramDisk(ID string, Host_memory_limit string, Zram_fraction float64, Max_zram_size string, Compression_algorithm string, Fs_type string, Swap_priority int, Mount_point string) ZramDiskConfiguration {
	return ZramDiskConfiguration{ID, Host_memory_limit, Zram_fraction, Max_zram_size, Compression_algorithm, Fs_type, Swap_priority, Mount_point}
}

func NewZramConfiguration(Name string, ZramDisks []ZramDiskConfiguration) ZramConfigurations {
	return ZramConfigurations{Name, ZramDisks}
}

func (zrd ZramDiskConfiguration) ToString() string {
	if zrd.Fs_type == SWAP {
		return fmt.Sprintf("[%s]\n%s\n%f\n%s\n%s\n%s\n%d\n", zrd.ID, zrd.Host_memory_limit, zrd.Zram_fraction, zrd.Max_zram_size, zrd.Compression_algorithm, zrd.Fs_type, zrd.Swap_priority)
	}
	return fmt.Sprintf("[%s]\n%s\n%f\n%s\n%s\n%s\n%s\n", zrd.ID, zrd.Host_memory_limit, zrd.Zram_fraction, zrd.Max_zram_size, zrd.Compression_algorithm, zrd.Fs_type, zrd.Mount_point)
}
