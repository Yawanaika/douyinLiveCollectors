package main

import "douyinLiveCollectors/collectors"

func main() {

	collectors.NewLiveViewer(766678346504).Start()

	select {}
}
