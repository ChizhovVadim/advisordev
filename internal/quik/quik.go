package quik

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"net"
	"strconv"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Не потокобезопасный
type QuikService struct {
	id       int64
	transId  int64
	mainConn net.Conn
	reader   *bufio.Reader
}

func InitConnection(port int) (net.Conn, error) {
	if port == 0 {
		port = 34130
	}
	conn, err := net.Dial("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewQuikService(
	mainConn net.Conn,
) *QuikService {
	return &QuikService{
		mainConn: mainConn,
		reader:   bufio.NewReader(transform.NewReader(mainConn, charmap.Windows1251.NewDecoder())),
		id:       1,
		transId:  calculateStartTransId(),
	}
}

func (quik *QuikService) ExecuteQuery(
	command string,
	request interface{},
	response interface{},
) error {
	var r = RequestJson{
		Id:          quik.id,
		Command:     command,
		CreatedTime: timeToQuikTime(time.Now()),
		Data:        request,
	}
	quik.id += 1
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = quik.mainConn.Write(b)
	if err != nil {
		return err
	}
	_, err = quik.mainConn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	incoming, err := quik.reader.ReadString('\n')
	if err != nil {
		return err
	}
	var responseJson ResponseJson
	err = json.Unmarshal([]byte(incoming), &responseJson)
	if err != nil {
		return err
	}
	if responseJson.LuaError != "" {
		return errors.New(responseJson.LuaError)
	}
	if responseJson.Data != nil && response != nil {
		var err = json.Unmarshal(*responseJson.Data, response)
		if err != nil {
			return err
		}
		type nullable interface {
			Valid() bool
			SetValid(v bool)
		}
		if nullable, ok := response.(nullable); ok {
			nullable.SetValid(true)
		}
	}
	return nil
}

func calculateStartTransId() int64 {
	var hour, min, sec = time.Now().Clock()
	return 60*(60*int64(hour)+int64(min)) + int64(sec)
}

func timeToQuikTime(time time.Time) int64 {
	return time.UnixNano() / 1000
}

func QuikCallbacks(
	r io.Reader,
) iter.Seq2[CallbackJson, error] {
	return func(yield func(CallbackJson, error) bool) {
		reader := bufio.NewReader(transform.NewReader(r, charmap.Windows1251.NewDecoder()))
		for {
			incoming, err := reader.ReadString('\n')
			if err != nil {
				yield(CallbackJson{}, err)
				return
			}
			var callbackJson CallbackJson
			err = json.Unmarshal([]byte(incoming), &callbackJson)
			if err != nil {
				yield(CallbackJson{}, err)
				return
			}
			if !yield(callbackJson, nil) {
				return
			}
		}
	}
}
