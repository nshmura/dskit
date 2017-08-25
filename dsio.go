package main

import (
	"errors"
	"os"
	"strings"

	"github.com/nshmura/dsio/action"
	"github.com/nshmura/dsio/core"
	"github.com/urfave/cli"
)

const (
	version = "0.2.0"

	// the number of entities to output at once
	defaultPageSize = 50

	// max page size
	maxPageSize = 1000
)

var (
	FlagServiceAccoutFile = cli.StringFlag{
		Name:   "key-file",
		Usage:  "name of GCP service account file.",
		EnvVar: "DSIO_KEY_FILE",
	}

	FlagProjectID = cli.StringFlag{
		Name:   "project-id",
		Usage:  "Project ID of GCP.",
		EnvVar: "DSIO_PROJECT_ID",
	}

	FlagVerbose = cli.BoolFlag{
		Name:  "verbose, v",
		Usage: "Make the operation more talkative.",
	}

	FlagNoColor = cli.BoolFlag{
		Name:  "no-color",
		Usage: "Disable color output.",
	}

	FlagNamespace = cli.StringFlag{
		Name:  "namespace, n",
		Usage: "namespace of entities.",
	}
)

func main() {
	app := cli.NewApp()

	cli.VersionFlag = cli.BoolFlag{
		Name:  "version, V",
		Usage: "Show version number and quit",
	}

	app.Name = "dsio"
	app.Usage = "A command line tool for Google Cloud Datastore."
	app.Version = version

	app.Commands = []cli.Command{
		{
			Name:      "upsert",
			Usage:     "Bulk-upsert entities into Datastore.",
			ArgsUsage: "filename",
			Flags: []cli.Flag{
				FlagNamespace,
				cli.StringFlag{
					Name:  "kind, k",
					Usage: "Name of destination kind.",
				},
				cli.StringFlag{
					Name:  "format, f",
					Usage: "Format of input file. <yaml|csv|tcv>.",
				},
				cli.BoolFlag{
					Name:  "dry-run",
					Usage: "Skip Datastore operations.",
				},
				cli.IntFlag{
					Name:  "batch-size",
					Value: 500,
					Usage: "The number of entities per one multi upsert operation. batch-size should be smaller than 500.",
				},
				FlagServiceAccoutFile,
				FlagProjectID,
				FlagVerbose,
				FlagNoColor,
			},
			Action: func(c *cli.Context) error {
				args := c.Args()
				if l := len(args); l == 0 {
					core.Error("Filename is not specified")
					return nil

				} else if l > 1 {
					core.Error("Too many args")
					return nil
				}
				filename := args[0]

				ctx := core.SetContext(c)
				ctx.PrintContext()

				action.Upsert(ctx, filename, c.String("kind"), c.String("format"), c.Int("batch-size"))

				return nil
			},
		},
		{
			Name:      "query",
			Usage:     "Execute a query.",
			ArgsUsage: `"[<gql_query>]"`,
			Flags: []cli.Flag{
				FlagNamespace,
				cli.StringFlag{
					Name:  "output, o",
					Usage: "Output filename. Entities are outputed into this file.",
				},
				cli.StringFlag{
					Name:  "format, f",
					Value: "yaml",
					Usage: "Format of output. <yaml|csv|tcv>.",
				},
				cli.StringFlag{
					Name:  "style, s",
					Value: "scheme",
					Usage: "Style of output. <scheme|direct|auto>.",
				},
				cli.IntFlag{
					Name:  "page-size",
					Value: 50,
					Usage: "Number of entities to output at once.",
				},
				FlagServiceAccoutFile,
				FlagProjectID,
				FlagVerbose,
				FlagNoColor,
			},
			Action: func(c *cli.Context) error {
				query := strings.Join(c.Args(), " ")

				var format = c.String("format")
				switch format {
				case core.FormatCSV, core.FormatTSV, core.FormatYAML:
				// ok
				case "":
					format = core.FormatYAML
				default:
					core.Errorf("Format should be yaml, csv or tsv")
					return nil
				}

				style, err := getTypeStyle(c.String("style"))
				if err != nil {
					core.Error(err)
					return nil
				}

				pageSize := c.Int("page-size")
				if pageSize == 0 {
					pageSize = defaultPageSize

				} else if pageSize > maxPageSize {
					return core.Errorf("Too large page size:%v", pageSize)
				}

				ctx := core.SetContext(c)
				ctx.PrintContext()
				action.Query(ctx, query, format, style, c.String("output"), pageSize)
				return nil
			},
		},
	}

	app.Run(os.Args)
}

func getTypeStyle(style string) (core.TypeStyle, error) {
	switch style {
	case string(core.StyleScheme), string(core.StyleDirect), string(core.StyleAuto):
		return core.TypeStyle(style), nil
	case "":
		return core.StyleScheme, nil
	default:
		return core.TypeStyle(""), errors.New("Format should be one of scheme, direct, auto")
	}

}
