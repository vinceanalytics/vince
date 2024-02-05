package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/apache/arrow/go/v15/arrow/ipc"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/blevesearch/vellum"
	"github.com/vinceanalytics/staples/staples/staples"
	"gopkg.in/yaml.v2"
)

type Model struct {
	// UserAgent   string
	IsBot int64
	// BotName     string
	// BotCategory string

	OSName    string
	OSVersion string
	// OSPlatform string

	ClientName    string
	ClientType    string
	ClientVersion string
	// ClientEngine        string
	// ClientEngineVersion string

	// DeviceType  string
	// DeviceBrand string
	// DeviceModel string

	// OsFamily      string
	// BrowserFamily string
}

type Fixture struct {
	UserAgent     string  `yaml:"user_agent"`
	Bot           *Bot    `yaml:"bot"`
	Os            *Os     `yaml:"os"`
	Client        *Client `yaml:"client"`
	Device        *Device `yaml:"device"`
	OsFamily      string  `yaml:"os_family"`
	BrowserFamily string  `yaml:"browser_family"`
}

func (f *Fixture) Model() (m *Model) {
	m = &Model{
		// UserAgent:     f.UserAgent,
		// OsFamily:      f.OsFamily,
		// BrowserFamily: f.BrowserFamily,
	}

	if v := f.Bot; v != nil {
		m.IsBot = 1
		// m.BotName = v.Name
		// m.BotCategory = v.Category
	}
	if v := f.Os; v != nil {
		m.OSName = v.o.Name
		m.OSVersion = v.o.Version
		// m.OSPlatform = v.o.Platform
	}
	if v := f.Client; v != nil {
		m.ClientName = v.Name
		m.ClientType = v.Type
		m.ClientVersion = v.Version
		// m.ClientEngine = v.Engine
		// m.ClientEngineVersion = v.EngineVersion
	}
	if v := f.Device; v != nil {
		// m.DeviceType = v.Type
		// m.DeviceBrand = v.Brand
		// m.DeviceModel = v.Model
	}
	return
}

func (f *Fixture) Merge(o *Fixture) {
	if o.Bot != nil {
		f.Bot = o.Bot
	}
	if o.Os != nil {
		f.Os = o.Os
	}
	if o.Client != nil {
		f.Client = o.Client
	}
	if o.Device != nil {
		f.Device = o.Device
	}
	if o.OsFamily != "" {
		f.OsFamily = o.OsFamily
	}
	if o.BrowserFamily != "" {
		f.BrowserFamily = o.BrowserFamily
	}
}

type Bot struct {
	Name     string `yaml:"name"`
	Category string `yaml:"category"`
}

type Os struct {
	o OsImpl
}
type OsImpl struct {
	Name     string `yaml:"name"`
	Version  string `yaml:"version"`
	Platform string `yaml:"platform"`
}

var _ yaml.Unmarshaler = (*Os)(nil)

func (o *Os) UnmarshalYAML(unmarshal func(interface{}) error) error {
	unmarshal(&o.o)
	return nil
}

type Client struct {
	Name          string `yaml:"name"`
	Type          string `yaml:"type"`
	Version       string `yaml:"version"`
	Engine        string `yaml:"engine"`
	EngineVersion string `yaml:"engine_version"`
}

type Device struct {
	Type  string `yaml:"type"`
	Brand string `yaml:"brand"`
	Model string `yaml:"model"`
}

func main() {
	flag.Parse()
	root := flag.Arg(0)
	files, err := os.ReadDir(root)
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]*Fixture)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if filepath.Ext(file.Name()) != ".yml" {
			continue
		}

		o := readUA(filepath.Join(root, file.Name()))
		for _, f := range o {
			g, ok := m[f.UserAgent]
			if ok {
				g.Merge(f)
				continue
			}
			m[f.UserAgent] = f
		}
	}
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)
	o := staples.NewArrow[Model](memory.DefaultAllocator)
	for i := range names {
		o.Append(m[names[i]].Model())
	}
	r := o.NewRecord()
	var b bytes.Buffer
	w := ipc.NewWriter(&b, ipc.WithSchema(r.Schema()), ipc.WithZstd())
	err = w.Write(r)
	if err != nil {
		log.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("ua.arrow", b.Bytes(), 0600)

	b.Reset()
	fst, err := vellum.New(&b, nil)
	if err != nil {
		log.Fatal(err)
	}
	for i := range names {
		err = fst.Insert([]byte(names[i]), uint64(i))
		if err != nil {
			log.Fatal(err)
		}
	}
	err = fst.Close()
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("ua.fst", b.Bytes(), 0600)
}

func readUA(path string) (out []*Fixture) {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(f, &out)
	if err != nil {
		log.Fatal("failed to  decode ", path, err.Error())
	}
	return
}
