package app

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/moncho/dry/appui"
	"github.com/moncho/dry/ui"
	"github.com/nsf/termbox-go"
)

type imagesScreenEventHandler struct {
	baseEventHandler
	passingEvents bool
}

func (h *imagesScreenEventHandler) widget() appui.EventableWidget {
	return h.dry.widgetRegistry.ImageList
}

func (h *imagesScreenEventHandler) handle(event termbox.Event) {
	if h.passingEvents {
		h.eventChan <- event
		return
	}
	//Controls if the event has been handled by the first switch statement
	var handled bool
	var keepFocus bool

	handled, keepFocus = h.handleKeyEvent(event.Key)

	if !handled {
		handled, keepFocus = h.handleChEvent(event.Ch)
	}
	if handled {
		h.setFocus(keepFocus)
		if h.hasFocus() {
			refreshScreen()
		}
	} else {
		h.baseEventHandler.handle(event)
	}
}

func (h *imagesScreenEventHandler) handleKeyEvent(key termbox.Key) (bool, bool) {
	dry := h.dry
	screen := h.screen
	keepFocus := true
	handled := true
	switch key {
	case termbox.KeyF1: //sort
		h.dry.widgetRegistry.ImageList.Sort()
	case termbox.KeyF5: // refresh
		h.widget().Unmount()
	case termbox.KeyCtrlD: //remove dangling images
		dry.RemoveDanglingImages()
	case termbox.KeyCtrlE: //remove image
		rmImage := func(id string) error {
			dry.RemoveImage(id, false)
			return nil
		}
		if err := h.widget().OnEvent(rmImage); err != nil {
			dry.appmessage(
				fmt.Sprintf("Error removing image: %s", err.Error()))
		}

	case termbox.KeyCtrlF: //force remove image
		rmImage := func(id string) error {
			dry.RemoveImage(id, true)
			return nil
		}
		if err := h.widget().OnEvent(rmImage); err != nil {
			dry.appmessage(
				fmt.Sprintf("Error forcing image removal: %s", err.Error()))
		}
	case termbox.KeyEnter: //inspect image
		inspectImage := func(id string) error {
			network, err := h.dry.dockerDaemon.InspectImage(id)
			if err != nil {
				return err
			}
			keepFocus = false
			renderer := appui.NewJSONRenderer(network)
			go appui.Less(renderer, screen, h.eventChan, h.closeViewChan)
			return nil
		}
		if err := h.widget().OnEvent(inspectImage); err != nil {
			dry.appmessage(
				fmt.Sprintf("Error inspecting image: %s", err.Error()))
		}

	default:
		handled = false
	}
	return handled, keepFocus
}

func (h *imagesScreenEventHandler) handleChEvent(ch rune) (bool, bool) {
	dry := h.dry
	screen := h.screen
	keepFocus := true
	handled := true
	switch ch {
	case '2': //Ignore since dry is already on the images screen

	case 'i', 'I': //image history

		history := func(id string) error {
			history, err := h.dry.dockerDaemon.History(id)
			if err != nil {
				return err
			}
			keepFocus = false
			renderer := appui.NewDockerImageHistoryRenderer(history)
			go appui.Less(renderer, screen, h.eventChan, h.closeViewChan)
			return nil
		}
		if err := h.widget().OnEvent(history); err != nil {
			dry.appmessage(err.Error())
		}
	case 'r', 'R': //Run container
		runImage := func(id string) error {
			image, err := h.dry.dockerDaemon.ImageByID(id)
			if err != nil {
				return err
			}
			rw := appui.NewImageRunWidget(image)
			h.passingEvents = true
			dry.widgetRegistry.add(rw)
			go func(image types.ImageSummary) {
				events := ui.EventSource{
					Events: h.eventChan,
					EventHandledCallback: func(e termbox.Event) error {
						return refreshScreen()
					},
				}
				rw.OnFocus(events)
				dry.widgetRegistry.remove(rw)
				runCommand, canceled := rw.Text()
				h.passingEvents = false
				if canceled {
					return
				}
				if err := dry.dockerDaemon.RunImage(image, runCommand); err != nil {
					dry.appmessage(err.Error())
				}
			}(image)
			return nil
		}
		if err := h.widget().OnEvent(runImage); err != nil {
			dry.appmessage(
				fmt.Sprintf("Error running image: %s", err.Error()))
		}
	default:
		handled = false

	}
	return handled, keepFocus
}
