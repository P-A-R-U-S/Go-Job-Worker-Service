package main

import (
	"github.com/P-A-R-U-S/Go-Job-Worker-Service/pkg/proto"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	a := cli.NewApp()
	a.Name = "Jow Worker Client"
	a.Usage = "Connect to JobWorker Service to run arbitrary Linux command on remote hosts"
	a.Email = "ValentynPonomarenko@gmail.com"

	a.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "start",
			Usage: "jw start --client-cert <PATH_TO_CLIENT_CERT> --cpu 0.5 --memory 500000 --io 1000000 --c $(which date)",
		},
		cli.StringFlag{
			Name:  "status",
			Usage: "jw status --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
		cli.StringFlag{
			Name:  "stream",
			Usage: "jw stream --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
		cli.StringFlag{
			Name:  "stop",
			Usage: "jw stop --client-cert <PATH_TO_CLIENT_CERT> --id <UUID>",
		},
	}

	a.Action = func(c *cli.Context) error {
		var err error

		if len(c.Args()) == 0 || c.IsSet("h") || c.IsSet("help") {
			err = cli.ShowAppHelp(c)
		}

		if c.IsSet("start") {

		}

		if c.IsSet("status") {

		}

		if c.IsSet("stream") {

		}

		if c.IsSet("stop") {

		}
		return err
	}

	err := a.Run(os.Args)

	if err != nil {
		log.Fatal(err)
	}
}

func getGRPCClient() *proto.JobWorkerClient {

}
