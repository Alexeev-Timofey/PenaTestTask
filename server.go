package main

import (
	"log"
	"time"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/r3labs/sse/v2"
)

const REFRESH_DELAY = "1s"
const STREAM_NAME = "word"
const LISTEN_ADDR = "127.0.0.1:25565"

type ChaneWordRequestData struct {
	Word string `json:word binding:required`
}

type WordGenerator struct {
	word string
	delay time.Duration
	sse_server *sse.Server
}

func NewWordGenerator(w string, d string) (WordGenerator, error) {
	del, err := time.ParseDuration(d)
	if err != nil {
		log.Println("Fail to parse delay string")
	}
	ret := WordGenerator {
		word: w,
		delay: del,
		sse_server: sse.New(),
	}
	ret.sse_server.CreateStream(STREAM_NAME)
	return ret, nil
}

func (wg *WordGenerator) HandleChangeWord(c *gin.Context) {
	var rdata ChaneWordRequestData
	if err := c.BindJSON(rdata); err != nil {
		log.Println("Fail to parse JSON in change word request")
		return
	}
	
	wg.word = rdata.Word
}

func (wg *WordGenerator) HandleNewSSE(c *gin.Context) {
	new_req := c.Request // Может надо c.Request.Clone()?
	query_vals := new_req.URL.Query()

	query_vals.Add("stream", STREAM_NAME)
	new_req.URL.RawQuery = query_vals.Encode()
	
	wg.sse_server.ServeHTTP(c.Writer, new_req)
}

func (wg *WordGenerator) SendWordThead() {
	wg.sse_server.Publish(STREAM_NAME, &sse.Event{
		Data: []byte(wg.word),
	})

	time.AfterFunc(wg.delay, wg.SendWordThead)
}

func main() {
	router := gin.Default()
	generator, err := NewWordGenerator("test", REFRESH_DELAY)
	if err != nil {
		os.Exit(1)
	}

	router.GET("/listen", generator.HandleNewSSE)
	router.POST("/say", generator.HandleChangeWord)
	
	generator.SendWordThead()

	router.Run(LISTEN_ADDR)
}
