package main

import "douyinLiveCollectors/collectors"

func main() {

	collectors.NewLiveViewer(658471243708).Start()

	select {}
}
