package main

import (
	"context"
	"errors"
	"rh363/zrgui/zram"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/taigrr/systemctl"
)

// allowed option
var options = []zram.Option{
	zram.NewOption("option1", "10", []string{"10", "20"}),
	zram.NewOption("option2", "20", []string{"20", "30"}),
	zram.NewOption("option3", "30", []string{"30", "40"}),
}

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
	w.Resize(fyne.NewSize(300, 500))

	//bind option struct in to bindingitem format
	bindedOptions := binding.NewUntypedList()
	for _, option := range options {
		bindedOptions.Append(option)
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
	ItemMenu := widget.NewListWithData(
		bindedOptions,
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewSelectEntry([]string{}),
				widget.NewLabel(""),
			)
		},
		func(di binding.DataItem, co fyne.CanvasObject) {
			container := co.(*fyne.Container)
			optionLabel := container.Objects[0].(*widget.Label)
			selectEntry := container.Objects[1].(*widget.SelectEntry)

			option := zram.NewOptionFromDataItem(di)
			optionLabel.SetText(option.Name)
			selectEntry.SetText(option.State)
			selectEntry.SetOptions(option.Values)
		},
	)

	/*Applybtn
	apply configuration
	*/
	ApplyBtn := widget.NewButton("Apply", func() {
		//save config//
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
