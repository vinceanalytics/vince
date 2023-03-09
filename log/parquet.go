package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog"
	"github.com/segmentio/parquet-go"
	"github.com/urfave/cli/v2"
)

type Event struct {
	Timestamp time.Time         `parquet:"timestamp"`
	Level     string            `parquet:"level,dict,zstd"`
	Message   string            `parquet:"msg,dict,zstd"`
	Fields    map[string]string `parquet:"fields" parquet-key:",dict,zstd" parquet-value:",dict,zstd"`
}

type Columns struct {
	Timestamps []parquet.Value
	Levels     []parquet.Value
	Messages   []parquet.Value
	Keys       []parquet.Value
	Values     []parquet.Value
}

const format = "1/02 15:04:05"

func (c *Columns) Write(o io.Writer, duration time.Duration) {
	table := tablewriter.NewWriter(o)
	table.SetHeader([]string{"timestamp", "level", "key", "value", "msg"})
	table.SetCaption(true, fmt.Sprintf("elapsed %s", duration))
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t") // pad with tabs
	table.SetNoWhiteSpace(true)
	for i := range c.Timestamps {
		ls := make([]string, 5)
		ls[0] = time.Unix(0, c.Timestamps[i].Int64()).Format(format)
		ls[1] = c.Levels[i].String()
		ls[2] = c.Keys[i].String()
		ls[3] = c.Values[i].String()
		ls[4] = c.Messages[i].String()
		table.Append(ls)
	}
	table.Render()

}

const BUFFER_SIZE = 2 << 10

type Rotate struct {
	filename string
	f        *os.File
	sw       *parquet.SortingWriter[*Event]
	buf      []*Event
	mu       sync.Mutex
}

func NewRotate(path string) (*Rotate, error) {
	b := &Rotate{
		filename: filepath.Join(path, "error_log"),
		buf:      make([]*Event, 0, BUFFER_SIZE),
	}
	return b, b.open()
}

func (b *Rotate) Write(p []byte) (int, error) {
	return 0, nil
}

func (b *Rotate) Rotate() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.close()
	if err != nil {
		return err
	}
	return b.reopen()
}

func (b *Rotate) open() error {
	var err error
	b.f, err = os.Create(b.filename)
	if err != nil {
		return fmt.Errorf("failed to create new log file %s  %v", b.filename, err)
	}
	b.sw = parquet.NewSortingWriter[*Event](b.f, 4<<10,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Descending("timestamp"),
			),
		),
	)
	return nil
}

func (b *Rotate) reopen() error {
	var err error
	b.f, err = os.Create(b.filename)
	if err != nil {
		return fmt.Errorf("failed to create new log file %s  %v", b.filename, err)
	}
	b.sw.Reset(b.f)
	return nil
}

func (b *Rotate) Close() error {
	return b.close()
}

func (b *Rotate) close() error {
	err := b.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush messages %v", err)
	}
	err = b.sw.Close()
	if err != nil {
		return fmt.Errorf("failed to close parquet writer %v", err)
	}
	now := time.Now().UTC().Format(time.DateOnly)
	out := filepath.Join(filepath.Dir(b.filename), now+"_log.parquet")
	o, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("failed to create rotation file %s  %v", out, err)
	}
	_, err = b.f.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek log file %v", err)
	}
	_, err = o.ReadFrom(b.f)
	if err != nil {
		return fmt.Errorf("failed to copy rotation file %v", err)
	}
	err = b.f.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file writer %v", err)
	}
	return nil
}

func (b *Rotate) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.buf) == 0 {
		return nil
	}
	_, err := b.sw.Write(b.buf)
	if err != nil {
		return err
	}
	b.buf = b.buf[:0]
	return nil
}

func (b *Rotate) WriteLevel(level zerolog.Level, p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var evt map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err := d.Decode(&evt)
	if err != nil {
		return 0, err
	}
	e := &Event{
		Level: level.String(),
	}

	for k, v := range evt {
		switch k {
		case zerolog.LevelFieldName:
		case zerolog.MessageFieldName:
			e.Message = v.(string)
		case zerolog.TimestampFieldName:
			x := v.(json.Number)
			n, err := x.Int64()
			if err != nil {
				return 0, err
			}
			e.Timestamp = time.Unix(n, 0)
		default:
			if e.Fields == nil {
				e.Fields = make(map[string]string)
			}
			e.Fields[k] = fmt.Sprint(v)
		}
	}
	b.buf = append(b.buf, e)
	return len(p), nil
}

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "log",
		Usage: "process error logs  generated vince",
		Action: func(ctx *cli.Context) error {
			return analyze(ctx.Args().First())
		},
	}
}

func analyze(file string) error {
	start := time.Now()
	if file == "" {
		return nil
	}
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to ope file %s %v", file, err)
	}
	defer f.Close()
	stat, _ := f.Stat()

	r, err := parquet.OpenFile(f, stat.Size())
	if err != nil {
		return fmt.Errorf("failed to open parquet file %s %v", file, err)
	}
	result := Columns{}
	pages := make([]parquet.Pages, 5)
	for _, g := range r.RowGroups() {
		columns := g.ColumnChunks()
		for i := range columns {
			pages[i] = columns[i].Pages()
		}
		{
			for i, p := range pages {
				values, err := readPage(p)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
				}
				switch i {
				case 0:

					result.Timestamps = append(result.Timestamps, values...)
				case 1:
					result.Levels = append(result.Levels, values...)

				case 2:
					result.Messages = append(result.Messages, values...)

				case 3:
					result.Keys = append(result.Keys, values...)

				case 4:
					result.Values = append(result.Values, values...)
				}
			}
		}
		for _, p := range pages {
			p.Close()
		}
	}
	result.Write(os.Stdout, time.Since(start))
	return nil
}

func readPage(pages parquet.Pages) ([]parquet.Value, error) {
	page, err := pages.ReadPage()
	if err != nil {
		return nil, err
	}
	values := make([]parquet.Value, page.NumValues())
	_, err = page.Values().ReadValues(values)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return values, nil
		}
		return nil, err
	}
	return values, nil
}
