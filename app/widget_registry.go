package app

import (
	"github.com/moncho/dry/appui"
	"github.com/moncho/dry/appui/swarm"
	"github.com/moncho/dry/docker"
	"github.com/moncho/dry/ui"
	"github.com/moncho/dry/ui/termui"
)

//WidgetRegistry holds two sets of widgets:
// * those registered in the the registry when it was created, that
//   can be reused. These are the individually named widgets found on
//   this struct.
// * a list of widgets to be rendered on the next rendering.
type WidgetRegistry struct {
	ContainerList    *appui.ContainersWidget
	DiskUsage        *appui.DockerDiskUsageRenderer
	DockerInfo       *appui.DockerInfo
	ImageList        *appui.DockerImagesWidget
	Monitor          *appui.Monitor
	Networks         *appui.DockerNetworksWidget
	Nodes            *swarm.NodesWidget
	NodeTasks        *swarm.NodeTasksWidget
	ServiceTasks     *swarm.ServiceTasksWidget
	ServiceList      *swarm.ServicesWidget
	activeWidgets    map[string]termui.Widget
	widgetForViewMap map[viewMode]termui.Widget
}

//NewWidgetRegistry creates the WidgetCatalog
func NewWidgetRegistry(daemon docker.ContainerDaemon) *WidgetRegistry {
	di := appui.NewDockerInfo(daemon)
	di.SetX(0)
	di.SetY(1)
	di.SetWidth(ui.ActiveScreen.Dimensions.Width)
	w := WidgetRegistry{
		DockerInfo:    di,
		ContainerList: appui.NewContainersWidget(daemon, appui.MainScreenHeaderSize),
		ImageList:     appui.NewDockerImagesWidget(daemon, appui.MainScreenHeaderSize),
		DiskUsage:     appui.NewDockerDiskUsageRenderer(ui.ActiveScreen.Dimensions.Height),
		Monitor:       appui.NewMonitor(daemon, appui.MainScreenHeaderSize),
		Networks:      appui.NewDockerNetworksWidget(daemon, appui.MainScreenHeaderSize),
		Nodes:         swarm.NewNodesWidget(daemon, appui.MainScreenHeaderSize),
		NodeTasks:     swarm.NewNodeTasksWidget(daemon, appui.MainScreenHeaderSize),
		ServiceTasks:  swarm.NewServiceTasksWidget(daemon, appui.MainScreenHeaderSize),
		ServiceList:   swarm.NewServicesWidget(daemon, appui.MainScreenHeaderSize),
		activeWidgets: make(map[string]termui.Widget),
	}

	initWidgetForViewMap(&w)

	return &w
}

func (wr *WidgetRegistry) widgetForView(v viewMode) termui.Widget {
	return wr.widgetForViewMap[v]
}

func (wr *WidgetRegistry) add(w termui.Widget) {
	if err := w.Mount(); err == nil {
		wr.activeWidgets[w.Name()] = w
	}
}

func (wr *WidgetRegistry) remove(w termui.Widget) {
	if err := w.Unmount(); err == nil {
		delete(wr.activeWidgets, w.Name())
	}
}

func initWidgetForViewMap(wr *WidgetRegistry) {
	viewMap := make(map[viewMode]termui.Widget)
	viewMap[Main] = wr.ContainerList
	viewMap[Networks] = wr.Networks
	viewMap[Images] = wr.ImageList
	viewMap[Monitor] = wr.Monitor
	viewMap[Nodes] = wr.Nodes
	viewMap[Services] = wr.ServiceList
	//viewMap[ServiceTasks] = wr.ServiceTasks
	//viewMap[Tasks] = wr.NodeTasks
	wr.widgetForViewMap = viewMap

}
