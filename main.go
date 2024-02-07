package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"rh363/zrgui/zram"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/rh363/filemanager"
	"github.com/taigrr/systemctl"
)

// allowed option
var Configurations zram.ZramConfigurations
var ConfigurationPath string = "/etc/systemd/zram-generator.conf"

// service states
var servicesStates = map[bool]string{
	true:  "RUNNING",
	false: "DOWN",
}

// check zram service
func ZramIsActive() (bool, error) {
	status, err := systemctl.IsActive(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	if err != nil {
		return false, errors.New("service not found")
	}
	return status, nil
}

// systemctl start zram
func ZramOn() error {
	err := systemctl.Start(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

// systemctl stop zram
func ZramOff() error {
	err := systemctl.Stop(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

// systemctl restart zram
func ZramRestart() error {
	err := systemctl.Restart(context.TODO(), "systemd-zram-setup@zram0", systemctl.Options{UserMode: false})
	return err
}

// start end configfile position
type start_end struct {
	start int
	end   int
}

// actual zram service state
var ZramService bool

func main() {
	//check root privileges
	if os.Geteuid() != 0 {
		panic("ERROR: USER MUST BE ROOT TO RUN THIS PROGRAM")
	}

	//get current service state
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

	//read configuration file
	activeConfigurationsFile, err := filemanager.ReadFile(ConfigurationPath)
	if err != nil {
		panic(err.Error())
	}

	//initialize start end struct array
	var start_ends []start_end
	start := -1 //default start value
	/*
		trough every line in configuration file search for start line
		and end line of every disk configuration avaible and save it
		in the previusly declared start end array
	*/
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

	//initialize zramdisks struct
	var zRamDisks zram.ZramConfigurations
	/*
		for every configuration defined between previusly finded line range
		create a new disk configuration and add it to zramdisks struct
	*/
	for _, start_end := range start_ends {
		//create a zram disk
		var zRamDisk zram.ZramDiskConfiguration
		for i := start_end.start; i <= start_end.end; i++ { //for every line in start end range try to understand what it is
			switch {
			case strings.Contains(activeConfigurationsFile[i], "#"): //if it is a comment skip it
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
		zRamDisks.Disks = append(zRamDisks.Disks, zRamDisk) //save it to zramdisks struct
	}

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
		//get list lenght
		func() int {
			return len(zRamDisks.Disks)
		},
		//define list item struct
		func() fyne.CanvasObject {
			//border container
			return container.NewBorder(
				//top: zrdamdisk id
				widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				//bottom,left,right
				nil, nil, nil,
				//remains space: option content
				//rows container
				container.NewGridWithRows(
					7, //rows number
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                   // column number
						widget.NewLabel(""), //used label
						widget.NewLabel(""), //unused label
						widget.NewSelect([]string{}, func(s string) {}), // select camp
					),
					container.NewGridWithColumns( //column container
						3,                                 // column number
						widget.NewLabel(""),               //used label
						widget.NewLabel(""),               //unused label
						widget.NewSelectEntry([]string{}), //entry select camp
					),
				),
			)
		},
		//define item list logic and content
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			MainContainer := co.(*fyne.Container) //get container from canvas
			zRamDisk := &zRamDisks.Disks[lii]     //get zramdisk reference

			title := MainContainer.Objects[1].(*widget.Label) //get title label
			title.SetText(zRamDisk.ID)                        //set title

			SubContainerRows := MainContainer.Objects[0].(*fyne.Container) // get rows container
			//get every row container and assign it to an option
			host_memory_limit_row := SubContainerRows.Objects[0].(*fyne.Container)
			zram_fraction_row := SubContainerRows.Objects[1].(*fyne.Container)
			max_zram_size_row := SubContainerRows.Objects[2].(*fyne.Container)
			compression_algorithm_row := SubContainerRows.Objects[3].(*fyne.Container)
			fs_type_row := SubContainerRows.Objects[4].(*fyne.Container)
			swap_priority_row := SubContainerRows.Objects[5].(*fyne.Container)
			mount_point_row := SubContainerRows.Objects[6].(*fyne.Container)

			//from every row get entry and label
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

			HMLLabel.SetText("host-memory-limit:")           //set text label
			HMLEntry.SetOptions(zram.MEMORYLIMIT_LIST)       //set option avaible
			HMLEntry.SetSelected(zRamDisk.Host_memory_limit) //write actually option
			HMLEntry.OnChanged = func(s string) {            //if it change update zramdisk
				zRamDisk.Host_memory_limit = HMLEntry.Selected
			}

			ZRFLabel.SetText("zram-fraction:")                                             //set text label
			ZRFEntry.SetOptions(zram.FRACTION_LIST)                                        //set option avaible
			ZRFEntry.SetSelected(strconv.FormatFloat(zRamDisk.Zram_fraction, 'f', -1, 64)) //convert and write actually option
			ZRFEntry.OnChanged = func(s string) {                                          //if it change update zramdisk
				zRamDisk.Zram_fraction, _ = strconv.ParseFloat(ZRFEntry.Selected, 64)
			}

			MZRSLabel.SetText("max-zram-size:")                         //set text label
			MZRSEntry.SetOptions(zram.MAXZRAMSIZE_LIST)                 //set option avaible
			MZRSEntry.SetSelected(strconv.Itoa(zRamDisk.Max_zram_size)) //convert and write actually option
			MZRSEntry.OnChanged = func(s string) {                      //if it change update zramdisk
				zRamDisk.Max_zram_size, _ = strconv.Atoi(MZRSEntry.Selected)
			}

			CALabel.SetText("compression-algorithm:")           //set text label
			CAEntry.SetOptions(zram.COMPRESSION_LIST)           //set option avaible
			CAEntry.SetSelected(zRamDisk.Compression_algorithm) //write actually option
			CAEntry.OnChanged = func(s string) {                // if it change update zramdisk
				zRamDisk.Compression_algorithm = CAEntry.Selected
			}

			FSTLabel.SetText("fs-type:")           //set text label
			FSTEntry.SetOptions(zram.FS_LIST)      //set option avaible
			FSTEntry.SetSelected(zRamDisk.Fs_type) //write actually option
			if zRamDisk.Fs_type == zram.SWAP {     //if FS is SWAP disable mount point option
				MPEntry.Disable()
			} else { //else disable SWAP PRIORITY option
				SPEntry.Disable()
			}
			FSTEntry.OnChanged = func(s string) { //if it change update zramdisk
				zRamDisk.Fs_type = FSTEntry.Selected
				if zRamDisk.Fs_type == zram.SWAP { //if FS is SWAP disable mount point option and active swap priority option
					SPEntry.Enable()
					MPEntry.Disable()
				} else { //else disable swap priority option and active mount option
					MPEntry.Enable()
					SPEntry.Disable()
				}
			}

			SPLabel.SetText("swap-priority:")                         //set text label
			SPEntry.SetOptions(zram.SWAPPRIO_LIST)                    //set option avaible
			SPEntry.SetSelected(strconv.Itoa(zRamDisk.Swap_priority)) //write actually option
			SPEntry.OnChanged = func(s string) {                      //if it is changed update zramdisk
				zRamDisk.Swap_priority, _ = strconv.Atoi(SPEntry.Selected)
			}

			MPLabel.SetText("mount-point:")          //set text label
			MPEntry.SetOptions(zram.MOUNTPOINT_LIST) //set option avaible
			MPEntry.SetText(zRamDisk.Mount_point)    //write actually option
			MPEntry.OnChanged = func(s string) {     //if it is changed update zramdisk
				zRamDisk.Mount_point = MPEntry.Text
			}

		},
	)

	/*Applybtn
	apply configuration
	*/
	ApplyBtn := widget.NewButton("Apply", func() {
		os.Remove(ConfigurationPath)                                  //remove configuration file
		filemanager.WriteFile(ConfigurationPath, zRamDisks.ToWrite()) //write new configuration file
		if ZramService {                                              // if service is running
			if err := ZramRestart(); err != nil { //restart service, if service restart failed
				dial := dialog.NewError(errors.New("ERROR - CANT RESTART SERVICE,SYSTEMCTL OR OPTION ERROR"), w) //print an alert
				dial.Show()
			}
		}
	})

	/*StBtn
	start/stop services
	*/
	StBtn := widget.NewButton("", func() {})

	/*
		set footer container and his content
	*/
	Footer := container.NewGridWithColumns(2, StBtn, ApplyBtn)

	if ZramService { //check zramservice and set button text
		StBtn.SetText("Stop")
	} else {
		StBtn.SetText("Start")
	}

	StBtn.OnTapped = func() { //if start/stop button is touched
		switch ZramService { //check zram service
		case true: //if is running
			if err := ZramOff(); err != nil { //try to take it off,if an error is returned
				dial := dialog.NewError(errors.New("ERROR - CANT RESTART SERVICE,SYSTEMCTL OR OPTION ERROR"), w) //display an alert
				dial.Show()
			} else { //else
				ZramService = false                             //update chek var
				StBtn.Text = "Start"                            //update button text
				ServiceState.Text = servicesStates[ZramService] //update state in header
				Footer.Refresh()                                //refresh view
				Header.Refresh()                                //refresh view
			}
		case false: //if it isn't running
			if err := ZramOn(); err != nil { //try to take it on,if an error is returned
				dial := dialog.NewError(errors.New("ERROR - CANT RESTART SERVICE,SYSTEMCTL OR OPTION ERROR"), w) //display an alert
				dial.Show()
			} else {
				ZramService = true                              //update check var
				StBtn.Text = "Stop"                             //update button text
				ServiceState.Text = servicesStates[ZramService] //update state in header
				Footer.Refresh()                                //refresh view
				Header.Refresh()                                //refresh view
			}
		}
	}
	//assemble element in window
	w.SetContent( //set window content
		container.NewBorder( //create a border container
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
	w.ShowAndRun() //run gui
}
