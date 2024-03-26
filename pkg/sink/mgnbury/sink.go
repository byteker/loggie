package mgnbury

import (
    "fmt"
    "github.com/loggie-io/loggie/pkg/core/api"
    "github.com/loggie-io/loggie/pkg/core/log"
    "github.com/loggie-io/loggie/pkg/core/result"
    "github.com/loggie-io/loggie/pkg/pipeline"
    "github.com/loggie-io/loggie/pkg/sink/codec"
    "net/http"
    "strings"
)

const (
    Type = "mgnbury"
)

func init() {
    pipeline.Register(api.SINK, Type, makeSink)
}

func makeSink(info pipeline.Info) api.Component {
    return NewSink()
}

type Sink struct {
    name    string
    config  *Config
    codec   codec.Codec
    client  *http.Client
    pushUrl string
}

func NewSink() *Sink {
    return &Sink{
        config: &Config{},
    }
}

func (s *Sink) Category() api.Category {
    return api.SINK
}

func (s *Sink) Type() api.Type {
    return Type
}

func (s *Sink) Config() interface{} {
    return s.config
}

func (s *Sink) SetCodec(c codec.Codec) {
    s.codec = c
}

func (s *Sink) String() string {
    return fmt.Sprintf("%s/%s", api.SINK, Type)
}

func (s *Sink) Init(context api.Context) error {
    s.name = context.Name()
    s.client = &http.Client{}
    s.pushUrl = s.config.Addr
    return nil
}

func (s *Sink) Start() error {
    log.Info("%s start", s.String())
    return nil
}

func (s *Sink) Stop() {
}

func (s *Sink) Consume(batch api.Batch) api.Result {
    events := batch.Events()
    l := len(events)
    if l == 0 {
        return nil
    }

    dataList := make([]string, 0)
    for _, e := range events {
        // json encode
        out, err := s.codec.Encode(e)
        if err != nil {
            log.Warn("codec event error: %+v", err)
            continue
        }
        dataList = append(dataList, string(out))
    }
    ndjson := fmt.Sprintf("[%s]", strings.Join(dataList, ","))
    req, err := http.NewRequest("POST", s.pushUrl, strings.NewReader(ndjson))
    if err != nil {
        return result.Fail(err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        return result.Fail(err)
    }
    defer resp.Body.Close()
    return result.NewResult(api.SUCCESS)
}
