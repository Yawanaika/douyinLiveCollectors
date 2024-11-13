package app

import (
	"context"
	"douyinLiveCollectors/backend/common/collectors"
	"douyinLiveCollectors/backend/common/log"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"strconv"
)

var Logger = log.GetLogger()

// App struct
type App struct {
	lv        *collectors.LiveViewer
	connected bool
	currentID uint64
	ctx       context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Shutdown() {
	if a.lv != nil {
		a.lv.Stop()
	}
	a.lv = nil
	a.connected = false
	a.currentID = 0
}
func (a *App) Start(id uint64) string {
	if a.connected && a.currentID == id {
		return "连接已建立，不能重复连接"
	}

	if a.connected && a.currentID != id {
		a.Shutdown()
	}

	a.lv = collectors.NewLiveViewer(id)
	a.connected = true
	a.currentID = id

	Logger.SetLiveId(strconv.FormatUint(a.currentID, 10))
	//err = os.MkdirAll(logDir, os.ModePerm)
	//if err != nil {
	//	return fmt.Sprintf("创建日志目录失败: %v", err)
	//}
	//
	//filename := fmt.Sprintf("%d.log", id)
	//filePath := filepath.Join(logDir, filename)
	//
	//a.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	//if err != nil {
	//	return "日志文件创建失败"
	//}
	a.lv.Start()

	go func() {
		for output := range a.lv.Out {

			runtime.EventsEmit(a.ctx, "new-output", output.Result)

			//if a.logFile != nil {
			//	_, err := a.logFile.WriteString(output.Result + "\n")
			//	if err != nil {
			//		Logger.Infof("写入日志文件失败:", err.Error())
			//	}
			//}
		}
	}()

	return "连接成功"
}
