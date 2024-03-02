package load

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/auth"
	"github.com/vinceanalytics/vince/internal/ref"
	"github.com/vinceanalytics/vince/ua"
	"google.golang.org/protobuf/encoding/protojson"
)

var client = &http.Client{}

const apiPath = "/api/v1/event"

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "load",
		Usage: "Generates events and send them to vince instance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "vince",
				Usage: "URL to vince instance you want to generate load for",
				Value: "http://localhost:8080",
			},
			&cli.IntFlag{
				Name:  "limit",
				Value: 10,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			vince := c.String("vince")
			program := &Program{
				Agents:   Agents(),
				Referrer: Referrer(),
				Vince:    vince,
			}
			vm := goja.New()
			vm.Set("createSession", program.NewSession)
			vm.Set("limit", c.Int("limit"))
			o := vm.NewObject()
			o.Set("log", fmt.Println)
			vm.Set("console", o)
			var data []byte
			a := c.Args().First()
			if a == "" {
				return nil
			}
			data, err := os.ReadFile(a)
			if err != nil {
				return fmt.Errorf("failed reading js file %q %v", a, err)
			}
			_, err = vm.RunString(string(data))
			return err
		},
	}
}

func Agents() []string {
	return ua.Random(10)
}

func Referrer() []string {
	return ref.Random(10)
}

type Program struct {
	Referrer []string
	Agents   []string
	Vince    string
}

func (p *Program) NewSession(website string) (*Session, error) {
	u, err := url.Parse(website)
	if err != nil {
		return nil, err
	}
	domain, _, _ := strings.Cut(u.Host, ":")
	return &Session{
		Ua:       p.Agents[rand.Intn(len(p.Agents))],
		Referrer: p.Referrer[rand.Intn(len(p.Referrer))],
		Domain:   domain,
		Website:  website,
		Vince:    p.Vince,
	}, nil
}

type Session struct {
	Ua       string
	Referrer string
	Domain   string
	Website  string
	Vince    string
}

func (s *Session) SendDebug(name, path string) error {
	return s.send(name, path, true)
}

func (s *Session) Send(name, path string) error {
	return s.send(name, path, false)
}

func (s *Session) send(name, path string, dump bool) error {
	e := &v1.Event{
		N:  name,
		U:  s.Website + path,
		D:  s.Domain,
		Ua: s.Ua,
		Ip: "127.0.0.1",
		R:  s.Referrer,
	}
	q := make(url.Values)
	q.Add("site_id", s.Domain)
	data, _ := protojson.Marshal(e)
	req, err := http.NewRequest(http.MethodPost, s.Vince+apiPath+"?"+q.Encode(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", auth.CreateBasicAuth("staples", ""))
	if dump {
		b, _ := httputil.DumpRequestOut(req, true)
		fmt.Println(string(b))
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		data, _ = io.ReadAll(res.Body)
		return errors.New(string(data))
	}
	return nil
}
