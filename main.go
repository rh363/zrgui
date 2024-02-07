package main

import (
	"context"
	"errors"
	"fmt"
	"rh363/zrgui/zram"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/rh363/filemanager"
	"github.com/taigrr/systemctl"
)

// allowed option
var Configurations zram.ZramConfigurations
var ConfigurationPath string = "/etc/systemd/zram-generator.conf.bk"

// service states
var servicesStates = map[bool]string{
	true:  "RUNNING",
	false: "DOWN",
}

func ZramIsActive() (bool, error) {
	status, err := systemctl.IsActive(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	if err != nil {
		return false, errors.New("service not found")
	}
	return status, nil
}

func ZramOn() error {
	err := systemctl.Start(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

func ZramOff() error {
	err := systemctl.Stop(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

func ZramRestart() error {
	err := systemctl.Restart(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

type start_end struct {
	start int
	end   int
}

var ZramService bool

func main() {
	//current service state
	if check, err := ZramIsActive(); err != nil {
		panic(err.Error())
	} else {
		if check {
			ZramService = true
		} else {
			ZramService = false
		}
	}
	//start gui app
	a := app.New()
	//create gui window
	w := a.NewWindow("ZRAM GUI")
	//define window size
	w.Resize(fyne.NewSize(500, 700))

	activeConfigurationsFile, err := filemanager.ReadFile(ConfigurationPath)
	if err != nil {
		panic(err.Error())
	}
	//strings.Split(strings.Split(conf, "=")[1], " ")[1]

	var start_ends []start_end
	start := -1
	for i, line := range activeConfigurationsFile {
		if strings.Contains(line, "[") && start < 0 {
			fmt.Println("find start: " + line)
			start = i
		} else if (strings.Contains(line, "[")) && start >= 0 {
			fmt.Println("find end/start: " + line)
			start_ends = append(start_ends, start_end{start, i - 1})
			start = i
		} else if i == len(activeConfigurationsFile)-1 && start >= 0 {
			fmt.Println("find end: " + line)
			start_ends = append(start_ends, start_end{start, i})
		}
	}
	fmt.Println(start_ends)

	var zRamDisks []zram.ZramDiskConfiguration
	for _, start_end := range start_ends {
		var zRamDisk zram.ZramDiskConfiguration
		for i := start_end.start; i <= start_end.end; i++ {
			switch {
			case strings.Contains(activeConfigurationsFile[i], "#"):
			case strings.Contains(activeConfigurationsFile[i], "["):
				zRamDisk.ID = strings.Split(strings.Split(activeConfigurationsFile[i], "[")[1], "]")[0]
			case strings.Contains(activeConfigurationsFile[i], "host-memory-limit"):
				zRamDisk.Host_memory_limit = strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1]
			case strings.Contains(activeConfigurationsFile[i], "zram-fraction"):
				zRamDisk.Zram_fraction, err = strconv.ParseFloat(strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1], 64)
				if err != nil {
					panic("ERROR - ZRAM FRACTION INVALID")
				}
			case strings.Contains(activeConfigurationsFile[i], "max-zram-size"):
				zRamDisk.Max_zram_size, err = strconv.Atoi(strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1])
				if err != nil {
					panic("ERROR - ZRAM MAX SIZE INVALID")
				}
			case strings.Contains(activeConfigurationsFile[i], "compression-algorithm"):
				zRamDisk.Compression_algorithm = strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1]
			case strings.Contains(activeConfigurationsFile[i], "fs-type"):
				zRamDisk.Fs_type = strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1]
			case strings.Contains(activeConfigurationsFile[i], "swap-priority"):
				zRamDisk.Swap_priority, err = strconv.Atoi(strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1])
				if err != nil {
					panic("ERROR - SWAP PRIORITY INVALID")
				}
			case strings.Contains(activeConfigurationsFile[i], "mount-point"):
				zRamDisk.Mount_point = strings.Split(strings.Split(activeConfigurationsFile[i], "=")[1], " ")[1]
			}
		}
		zRamDisks = append(zRamDisks, zRamDisk)
	}
	fmt.Println(zRamDisks)

	//service state widget
	ServiceState := widget.NewLabelWithStyle(
		servicesStates[ZramService],
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	/*header container
	header --> ServiceState
	*/
	Header := container.NewCenter(ServiceState)

	/*ItemMenu container
	the body of our application, contain every option
	*/
	ItemMenu := widget.NewList(
		func() int {
			return len(zRamDisks)
		},
		func() fyne.CanvasObject {
			return container.NewBorder(
				widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), nil, nil, nil,
				container.NewGridWithRows(
					7,
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelect([]string{}, func(s string) {}),
					),
					container.NewGridWithColumns(
						3,
						widget.NewLabel(""),
						widget.NewLabel(""),
						widget.NewSelectEntry([]string{}),
					),
				),
			)
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			MainContainer := co.(*fyne.Container)
			zRamDisk := &zRamDisks[lii]

			title := MainContainer.Objects[1].(*widget.Label)
			title.SetText(zRamDisk.ID)

			SubContainerRows := MainContainer.Objects[0].(*fyne.Container)
			host_memory_limit_row := SubContainerRows.Objects[0].(*fyne.Container)
			zram_fraction_row := SubContainerRows.Objects[1].(*fyne.Container)
			max_zram_size_row := SubContainerRows.Objects[2].(*fyne.Container)
			compression_algorithm_row := SubContainerRows.Objects[3].(*fyne.Container)
			fs_type_row := SubContainerRows.Objects[4].(*fyne.Container)
			swap_priority_row := SubContainerRows.Objects[5].(*fyne.Container)
			mount_point_row := SubContainerRows.Objects[6].(*fyne.Container)

			HMLLabel := host_memory_limit_row.Objects[0].(*widget.Label)
			HMLEntry := host_memory_limit_row.Objects[2].(*widget.Select)
			ZRFLabel := zram_fraction_row.Objects[0].(*widget.Label)
			ZRFEntry := zram_fraction_row.Objects[2].(*widget.Select)
			MZRSLabel := max_zram_size_row.Objects[0].(*widget.Label)
			MZRSEntry := max_zram_size_row.Objects[2].(*widget.Select)
			CALabel := compression_algorithm_row.Objects[0].(*widget.Label)
			CAEntry := compression_algorithm_row.Objects[2].(*widget.Select)
			FSTLabel := fs_type_row.Objects[0].(*widget.Label)
			FSTEntry := fs_type_row.Objects[2].(*widget.Select)
			SPLabel := swap_priority_row.Objects[0].(*widget.Label)
			SPEntry := swap_priority_row.Objects[2].(*widget.Select)
			MPLabel := mount_point_row.Objects[0].(*widget.Label)
			MPEntry := mount_point_row.Objects[2].(*widget.SelectEntry)

			HMLLabel.SetText("host-memory-limit:")
			HMLEntry.SetOptions(zram.MEMORYLIMIT_LIST)
			HMLEntry.SetSelected(zRamDisk.Host_memory_limit)
			HMLEntry.OnChanged = func(s string) {
				zRamDisk.Host_memory_limit = HMLEntry.Selected
			}

			ZRFLabel.SetText("zram-fraction:")
			ZRFEntry.SetOptions(zram.FRACTION_LIST)
			ZRFEntry.SetSelected(strconv.FormatFloat(zRamDisk.Zram_fraction, 'f', -1, 64))
			ZRFEntry.OnChanged = func(s string) {
				zRamDisk.Zram_fraction, _ = strconv.ParseFloat(ZRFEntry.Selected, 64)
			}

			MZRSLabel.SetText("max-zram-size:")
			MZRSEntry.SetOptions(zram.MAXZRAMSIZE_LIST)
			MZRSEntry.SetSelected(strconv.Itoa(zRamDisk.Max_zram_size))
			MZRSEntry.OnChanged = func(s string) {
				zRamDisk.Max_zram_size, _ = strconv.Atoi(MZRSEntry.Selected)
			}

			CALabel.SetText("compression-algorithm:")
			CAEntry.SetOptions(zram.COMPRESSION_LIST)
			CAEntry.SetSelected(zRamDisk.Compression_algorithm)
			CAEntry.OnChanged = func(s string) {
				zRamDisk.Compression_algorithm = CAEntry.Selected
			}

			FSTLabel.SetText("fs-type:")
			FSTEntry.SetOptions(zram.FS_LIST)
			FSTEntry.SetSelected(zRamDisk.Fs_type)
			if zRamDisk.Fs_type == zram.SWAP {
				MPEntry.Disable()
			} else {
				SPEntry.Disable()
			}
			FSTEntry.OnChanged = func(s string) {
				zRamDisk.Fs_type = FSTEntry.Selected
				if zRamDisk.Fs_type == zram.SWAP {
					SPEntry.Enable()
					MPEntry.Disable()
				} else {
					MPEntry.Enable()
					SPEntry.Disable()
				}
			}

			SPLabel.SetText("swap-priority:")
			SPEntry.SetOptions(zram.SWAPPRIO_LIST)
			SPEntry.SetSelected(strconv.Itoa(zRamDisk.Swap_priority))
			SPEntry.OnChanged = func(s string) {
				zRamDisk.Swap_priority, _ = strconv.Atoi(SPEntry.Selected)
			}

			MPLabel.SetText("mount-point:")
			MPEntry.SetOptions(zram.MOUNTPOINT_LIST)
			MPEntry.SetText(zRamDisk.Mount_point)
			SPEntry.OnChanged = func(s string) {
				zRamDisk.Mount_point = MPEntry.Text
			}

		},
	)

	/*Applybtn
	apply configuration
	*/
	ApplyBtn := widget.NewButton("Apply", func() {
		fmt.Println(zRamDisks)
	})

	/*StBtn
	start/stop services
	*/
	StBtn := widget.NewButton("", func() {})

	Footer := container.NewGridWithColumns(2, StBtn, ApplyBtn)

	if ZramService {
		StBtn.SetText("Stop")
	} else {
		StBtn.SetText("Start")
	}

	StBtn.OnTapped = func() {
		switch ZramService {
		case true:
			ZramService = false
			StBtn.Text = "Start"
			ServiceState.Text = servicesStates[ZramService]
			if err := ZramOff(); err != nil {
				panic(err.Error())
			}
			Footer.Refresh()
			Header.Refresh()
		case false:
			ZramService = true
			if err := ZramOn(); err != nil {
				panic(err.Error())
			}
			StBtn.Text = "Stop"
			ServiceState.Text = servicesStates[ZramService]
			Footer.Refresh()
			Header.Refresh()
		}
	}

	w.SetContent(
		container.NewBorder(
			//top
			Header,
			//bottom
			Footer,
			//left
			nil,
			//right
			nil,
			ItemMenu,
		),
	)
	w.ShowAndRun()
}
