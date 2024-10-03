package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/ro2"
)

func crack() *cli.Command {
	return &cli.Command{
		Name:  "crack",
		Usage: "Cracks vince license key",
		Description: `Allows users to use vince without a valid license key.
		# vince crack /path/to/vince/data`,
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:  "expires",
				Usage: "Duration of the patched license",
				Value: 24 * time.Hour,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			db, err := ro2.Open(c.Args().First())
			if err != nil {
				return err
			}
			defer db.Close()
			email := "crack@vinceanalytics.com"
			valid := time.Now().UTC().Add(c.Duration("expires"))
			err = db.PatchLicense(&v1.License{
				Expiry: uint64(valid.UnixMilli()),
				Email:  email,
			})
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
			fmt.Fprintln(w, "VINCE_ADMIN_EMAIL\tExpires\t")
			fmt.Fprintf(w, "%s\t%s\t\n", email, valid)
			return w.Flush()
		},
	}
}
