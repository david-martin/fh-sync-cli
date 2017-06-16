package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"net/http"

	ui "github.com/gizak/termui"
	"github.com/urfave/cli"
)

type statsResponse struct {
	CPUUsage struct {
		Worker1 struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"worker-1"`
	} `json:"CPU usage"`
	RSSMemoryUsage struct {
		Worker1 struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"worker-1"`
	} `json:"RSS Memory Usage"`
	JobProcessTime struct {
		SyncWorker struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"sync_worker"`
	} `json:"Job Process Time"`
	JobQueueSize struct {
		SyncWorker struct {
			Current         int       `json:"current"`
			Max             int       `json:"max"`
			Min             int       `json:"min"`
			Average         int       `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"sync_worker"`
		PendingWorker struct {
			Current         int       `json:"current"`
			Max             int       `json:"max"`
			Min             int       `json:"min"`
			Average         int       `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"pending_worker"`
		AckWorker struct {
			Current         int       `json:"current"`
			Max             int       `json:"max"`
			Min             int       `json:"min"`
			Average         int       `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"ack_worker"`
	} `json:"Job Queue Size"`
	APIProcessTime struct {
		Sync struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"sync"`
	} `json:"API Process Time"`
	MongodbOperationTime struct {
		DoUpdateManyDatasetClients struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"doUpdateManyDatasetClients"`
		DoListDatasetClients struct {
			Current         string    `json:"current"`
			Max             string    `json:"max"`
			Min             string    `json:"min"`
			Average         string    `json:"average"`
			NumberOfRecords int       `json:"numberOfRecords"`
			From            time.Time `json:"from"`
			End             time.Time `json:"end"`
		} `json:"doListDatasetClients"`
	} `json:"Mongodb Operation Time"`
}

func main() {
	app := cli.NewApp()
	app.Name = "fh-sync-cli"
	app.Usage = "fh-sync-cli <statsURL>"
	app.Action = func(c *cli.Context) error {
		renderStats(c.Args().Get(0))
		return nil
	}

	app.Run(os.Args)
}

func renderStats(statsURL string) {

	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	worker1 := ui.NewLineChart()
	worker1.BorderLabel = "Worker 1 Memory Usage"
	worker1.Height = 12
	// worker1.X = 5
	// worker1.Y = 5
	var ps []float64

	worker1.Data = (func() []float64 {
		n := 30
		ps = make([]float64, n)
		for i := range ps {
			ps[i] = 0 // TODO: this should be a non value rather than 0 e.g. 0% CPU usage is different than unknown CPU usage
		}
		return ps
	})()

	text1 := ui.NewPar("Some Text")
	text1.Height = 12

	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, worker1)),
		ui.NewRow(
			ui.NewCol(6, 0, text1)))

	ui.Body.Align()

	ui.Render(ui.Body)

	var updateInProgress = false
	ui.Handle("/timer/1s", func(e ui.Event) {
		t := e.Data.(ui.EvtTimer)
		// t is a EvtTimer
		if t.Count%2 == 0 && !updateInProgress { // TODO: Is this the only/best way to do a 5 second timer? 5s doesn't seem to work
			updateInProgress = true

			text1.Text = "TIMER"
			ui.Render(ui.Body)
			stats, err := getStats(statsURL)
			if err != nil {
				// TODO: this could be less panicky as it may start working in the next timer loop
				panic(err)
			}
			text1.Text = fmt.Sprintf("Latest Val: %s", stats.RSSMemoryUsage.Worker1.Current)
			ui.Render(ui.Body)

			// layout := "2006-01-02T15:00:00.000Z"

			worker1.Data = (func() []float64 {
				ps = ps[1:]
				currentWithoutSuffix := stats.RSSMemoryUsage.Worker1.Current[:len(stats.RSSMemoryUsage.Worker1.Current)-2]
				i, err := strconv.ParseFloat(currentWithoutSuffix, 64)
				if err != nil {
					panic(err)
				}
				ps = append(ps, i)
				return ps
			})()

			text1.Text = fmt.Sprintf("Latest Val: %s %f", stats.RSSMemoryUsage.Worker1.Current, ps[len(ps)-1])
			ui.Render(ui.Body)
			updateInProgress = false

			// stats, err := getStats(statsURL)
			// if err != nil {
			// 	// TODO: this could be less panicky as it may start working in the next timer loop
			// 	panic(err)
			// }

			// fmt.Printf("%v+", stats)
		}
	})

	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ui.Body.Width = ui.TermWidth()
		ui.Body.Align()
		ui.Clear()
		ui.Render(ui.Body)
	})

	// handle key q pressing
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		// press q to quit
		ui.StopLoop()
	})

	ui.Loop()

}

func getStats(statsURL string) (statsResponse, error) {
	res, err := http.Get(statsURL)
	if err != nil {
		return statsResponse{}, err
	}
	defer res.Body.Close()

	var body statsResponse

	json.NewDecoder(res.Body).Decode(&body)

	return body, nil
}
