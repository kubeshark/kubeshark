package redis

import (
	"bufio"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	askPrefix         = "ASK "
	movedPrefix       = "MOVED "
	clusterDownPrefix = "CLUSTERDOWN "
	busyPrefix        = "BUSY "
	noscriptPrefix    = "NOSCRIPT "

	dollarByte        = '$'
	asteriskByte      = '*'
	plusByte          = '+'
	minusByte         = '-'
	colonByte         = ':'
	notApplicableByte = '0'
)

// receive message from redis
type RedisInputStream struct {
	*bufio.Reader
	Buf   []byte
	count int
	limit int
}

func (r *RedisInputStream) readByte() (byte, error) {
	err := r.ensureFill()
	if err != nil {
		return 0, err
	}
	ret := r.Buf[r.count]
	r.count++
	return ret, nil
}

func (r *RedisInputStream) ensureFill() error {
	if r.count < r.limit {
		return nil
	}
	var err error
	r.limit, err = r.Read(r.Buf)
	if err != nil {
		return newConnectError(err.Error())
	}
	r.count = 0
	if r.limit == -1 {
		return newConnectError("Unexpected end of stream")
	}
	return nil
}

func (r *RedisInputStream) readLine() (string, error) {
	buf := ""
	for {
		err := r.ensureFill()
		if err != nil {
			return "", err
		}
		b := r.Buf[r.count]
		r.count++
		if b == '\r' {
			err := r.ensureFill()
			if err != nil {
				return "", err
			}
			c := r.Buf[r.count]
			r.count++
			if c == '\n' {
				break
			}
			buf += string(b)
			buf += string(c)
		} else {
			buf += string(b)
		}
	}
	if buf == "" {
		return "", newConnectError("It seems like server has closed the connection.")
	}
	return buf, nil
}

func (r *RedisInputStream) readLineBytes() ([]byte, error) {
	err := r.ensureFill()
	if err != nil {
		return nil, err
	}
	pos := r.count
	buf := r.Buf
	for {
		if pos == r.limit {
			return r.readLineBytesSlowly()
		}
		p := buf[pos]
		pos++
		if p == '\r' {
			if pos == r.limit {
				return r.readLineBytesSlowly()
			}
			p := buf[pos]
			pos++
			if p == '\n' {
				break
			}
		}
	}
	N := pos - r.count - 2
	line := make([]byte, N)
	j := 0
	for i := r.count; i <= N; i++ {
		line[j] = buf[i]
		j++
	}
	r.count = pos
	return line, nil
}

func (r *RedisInputStream) readLineBytesSlowly() ([]byte, error) {
	buf := make([]byte, 0)
	for {
		err := r.ensureFill()
		if err != nil {
			return nil, err
		}
		b := r.Buf[r.count]
		r.count++
		if b == 'r' {
			err := r.ensureFill()
			if err != nil {
				return nil, err
			}
			c := r.Buf[r.count]
			r.count++
			if c == '\n' {
				break
			}
			buf = append(buf, b)
			buf = append(buf, c)
		} else {
			buf = append(buf, b)
		}
	}
	return buf, nil
}

func (r *RedisInputStream) readIntCrLf() (int64, error) {
	err := r.ensureFill()
	if err != nil {
		return 0, err
	}
	buf := r.Buf
	isNeg := false
	if buf[r.count] == '-' {
		isNeg = true
	}
	if isNeg {
		r.count++
	}
	value := int64(0)
	for {
		err := r.ensureFill()
		if err != nil {
			return 0, err
		}
		b := buf[r.count]
		r.count++
		if b == '\r' {
			err := r.ensureFill()
			if err != nil {
				return 0, err
			}
			c := buf[r.count]
			r.count++
			if c != '\n' {
				return 0, newConnectError("Unexpected character!")
			}
			break
		} else {
			value = value*10 + int64(b) - int64('0')
		}
	}
	if isNeg {
		return -value, nil
	}
	return value, nil
}

type RedisProtocol struct {
	is *RedisInputStream
}

func NewProtocol(is *RedisInputStream) *RedisProtocol {
	return &RedisProtocol{
		is: is,
	}
}

func (p *RedisProtocol) Read() (packet *RedisPacket, err error) {
	x, r, err := p.process()
	if err != nil {
		return
	}
	packet = &RedisPacket{}
	packet.Type = r

	switch v := x.(type) {
	case []interface{}:
		array := v
		if len(array) > 0 {
			switch array[0].(type) {
			case []uint8:
				packet.Command = RedisCommand(strings.ToUpper(string(array[0].([]uint8))))
				if len(array) > 1 {
					switch array[1].(type) {
					case []uint8:
						packet.Key = string(array[1].([]uint8))
					case int64:
						packet.Key = fmt.Sprintf("%d", array[1].(int64))
					}
				}
				if len(array) > 2 {
					switch array[2].(type) {
					case []uint8:
						packet.Value = string(array[2].([]uint8))
					case int64:
						packet.Value = fmt.Sprintf("%d", array[2].(int64))
					}
				}
				if len(array) > 3 {
					packet.Value = fmt.Sprintf("[%s", packet.Value)
					for _, item := range array[3:] {
						switch j := item.(type) {
						case []uint8:
							packet.Value = fmt.Sprintf("%s, %s", packet.Value, j)
						case int64:
							packet.Value = fmt.Sprintf("%s, %d", packet.Value, j)
						}
					}
					packet.Value = strings.TrimSuffix(packet.Value, ", ")
					packet.Value = fmt.Sprintf("%s]", packet.Value)
				}
			default:
				msg := fmt.Sprintf("Unrecognized element in Redis array: %v", reflect.TypeOf(array[0]))
				err = errors.New(msg)
				return
			}
		}
	case []uint8:
		val := string(v)
		if packet.Type == types[plusByte] {
			packet.Keyword = RedisKeyword(strings.ToUpper(val))
			if !isValidRedisKeyword(keywords, packet.Keyword) {
				err = fmt.Errorf("Unrecognized keyword: %s", string(packet.Command))
				return
			}
		} else {
			packet.Value = val
		}
	case string:
		packet.Value = v
	case int64:
		packet.Value = fmt.Sprintf("%d", v)
	default:
		msg := fmt.Sprintf("Unrecognized Redis data type: %v", reflect.TypeOf(x))
		err = errors.New(msg)
		return
	}

	if packet.Command != "" {
		if !isValidRedisCommand(commands, packet.Command) {
			err = fmt.Errorf("Unrecognized command: %s", string(packet.Command))
			return
		}
	}

	return
}

func (p *RedisProtocol) process() (v interface{}, r RedisType, err error) {
	b, err := p.is.readByte()
	if err != nil {
		return nil, types[notApplicableByte], newConnectError(err.Error())
	}
	switch b {
	case plusByte:
		v, err = p.processSimpleString()
		r = types[plusByte]
		return
	case dollarByte:
		v, err = p.processBulkString()
		r = types[dollarByte]
		return
	case asteriskByte:
		v, err = p.processArray()
		r = types[asteriskByte]
		return
	case colonByte:
		v, err = p.processInteger()
		r = types[colonByte]
		return
	case minusByte:
		v, err = p.processError()
		r = types[minusByte]
		return
	default:
		return nil, types[notApplicableByte], newConnectError(fmt.Sprintf("Unknown reply: %b", b))
	}
}

func (p *RedisProtocol) processSimpleString() ([]byte, error) {
	return p.is.readLineBytes()
}

func (p *RedisProtocol) processBulkString() ([]byte, error) {
	l, err := p.is.readIntCrLf()
	if err != nil {
		return nil, newConnectError(err.Error())
	}
	if l == -1 {
		return nil, nil
	}
	line := make([]byte, 0)
	for {
		err := p.is.ensureFill()
		if err != nil {
			return nil, err
		}
		b := p.is.Buf[p.is.count]
		p.is.count++
		if b == '\r' {
			err := p.is.ensureFill()
			if err != nil {
				return nil, err
			}
			c := p.is.Buf[p.is.count]
			p.is.count++
			if c != '\n' {
				return nil, newConnectError("Unexpected character!")
			}
			break
		} else {
			line = append(line, b)
		}
	}
	return line, nil
}

func (p *RedisProtocol) processArray() ([]interface{}, error) {
	l, err := p.is.readIntCrLf()
	if err != nil {
		return nil, newConnectError(err.Error())
	}
	if l == -1 {
		return nil, nil
	}
	ret := make([]interface{}, 0)
	for i := 0; i < int(l); i++ {
		if obj, _, err := p.process(); err != nil {
			ret = append(ret, err)
		} else {
			ret = append(ret, obj)
		}
	}
	return ret, nil
}

func (p *RedisProtocol) processInteger() (int64, error) {
	return p.is.readIntCrLf()
}

func (p *RedisProtocol) processError() (interface{}, error) {
	msg, err := p.is.readLine()
	if err != nil {
		return nil, newConnectError(err.Error())
	}
	if strings.HasPrefix(msg, movedPrefix) {
		host, port, slot, err := p.parseTargetHostAndSlot(msg)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("MovedDataError: %s host: %s port: %d slot: %d", msg, host, port, slot), nil
	} else if strings.HasPrefix(msg, askPrefix) {
		host, port, slot, err := p.parseTargetHostAndSlot(msg)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("AskDataError: %s host: %s port: %d slot: %d", msg, host, port, slot), nil
	} else if strings.HasPrefix(msg, clusterDownPrefix) {
		return fmt.Sprintf("ClusterError: %s", msg), nil
	} else if strings.HasPrefix(msg, busyPrefix) {
		return fmt.Sprintf("BusyError: %s", msg), nil
	} else if strings.HasPrefix(msg, noscriptPrefix) {
		return fmt.Sprintf("NoScriptError: %s", msg), nil
	}
	return fmt.Sprintf("DataError: %s", msg), nil
}

func (p *RedisProtocol) parseTargetHostAndSlot(clusterRedirectResponse string) (host string, po int, slot int, err error) {
	arr := strings.Split(clusterRedirectResponse, " ")
	host, port := p.extractParts(arr[2])
	slot, _ = strconv.Atoi(arr[1])
	po, err = strconv.Atoi(port)
	return
}

func (p *RedisProtocol) extractParts(from string) (string, string) {
	idx := strings.LastIndex(from, ":")
	host := from
	if idx != -1 {
		host = from[0:idx]
	}
	port := ""
	if idx != -1 {
		port = from[idx+1:]
	}
	return host, port
}
