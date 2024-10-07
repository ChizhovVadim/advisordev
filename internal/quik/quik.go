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
	id      int64
	transId int64
	reader  *bufio.Reader
	writer  *transform.Writer
}

func InitConnection(port int) (net.Conn, error) {
	return net.Dial("tcp", "localhost:"+strconv.Itoa(port))
}

func NewQuikService(
	mainConn net.Conn,
) *QuikService {
	var quikCharmap = charmap.Windows1251
	return &QuikService{
		reader:  bufio.NewReader(transform.NewReader(mainConn, quikCharmap.NewDecoder())),
		writer:  transform.NewWriter(mainConn, quikCharmap.NewEncoder()),
		id:      1,
		transId: calculateStartTransId(),
	}
}

func (quik *QuikService) ExecuteQueryRaw(
	command string,
	request any,
	response *ResponseJson,
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
	_, err = quik.writer.Write(b)
	if err != nil {
		return err
	}
	//TODO b=append(b, "\r\n")
	_, err = quik.writer.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	incoming, err := quik.reader.ReadString('\n')
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(incoming), response)
	if err != nil {
		return err
	}
	if response.LuaError != "" {
		return errors.New(response.LuaError)
	}
	return nil
}

func (quik *QuikService) ExecuteQueryDynamic(
	command string,
	request any,
) (any, error) {
	var resp ResponseJson
	var err = quik.ExecuteQueryRaw(command, request, &resp)
	if err != nil {
		return nil, err
	}
	var res any
	if resp.Data != nil {
		err = json.Unmarshal(*resp.Data, &res)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (quik *QuikService) ExecuteQuery(
	command string,
	request any,
	response any,
) error {
	var resp ResponseJson
	var err = quik.ExecuteQueryRaw(command, request, &resp)
	if err != nil {
		return err
	}
	if resp.Data != nil && response != nil {
		var err = json.Unmarshal(*resp.Data, response)
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
