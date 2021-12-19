// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

/* GENERATED FILE - DO NOT EDIT */
/* Rebuild from the spec/gen.go tool */

package main

import (
	"encoding/binary"
	"io"
)

// Error codes that can be sent from the server during a connection or
// channel exception or used by the client to indicate a class of error like
// ErrCredentials.  The text of the error is likely more interesting than
// these constants.
const (
	frameMethod                 = 1
	frameHeader                 = 2
	frameBody                   = 3
	frameHeartbeat              = 8
	frameMinSize                = 4096
	frameEnd                    = 206
	replySuccess                = 200
	ContentTooLarge             = 311
	NoRoute                     = 312
	NoConsumers                 = 313
	ConnectionForced            = 320
	InvalidPath                 = 402
	AccessRefused               = 403
	NotFound                    = 404
	ResourceLocked              = 405
	PreconditionFailed          = 406
	FrameError                  = 501
	SyntaxError                 = 502
	CommandInvalid              = 503
	ChannelError                = 504
	UnexpectedFrame             = 505
	ResourceError               = 506
	NotAllowed                  = 530
	NotImplemented              = 540
	InternalError               = 541
	MaxSizeError                = 551
	MaxHeaderFrameSizeError     = 552
	BadMethodFrameUnknownMethod = 601
	BadMethodFrameUnknownClass  = 602
)

func isSoftExceptionCode(code int) bool {
	switch code {
	case 311:
		return true
	case 312:
		return true
	case 313:
		return true
	case 403:
		return true
	case 404:
		return true
	case 405:
		return true
	case 406:
		return true

	}
	return false
}

type ConnectionStart struct {
	VersionMajor     byte   `json:"versionMajor"`
	VersionMinor     byte   `json:"versionMinor"`
	ServerProperties Table  `json:"serverProperties"`
	Mechanisms       string `json:"mechanisms"`
	Locales          string `json:"locales"`
}

func (msg *ConnectionStart) id() (uint16, uint16) {
	return 10, 10
}

func (msg *ConnectionStart) wait() bool {
	return true
}

func (msg *ConnectionStart) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.VersionMinor); err != nil {
		return
	}

	if err = writeTable(w, msg.ServerProperties); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Mechanisms); err != nil {
		return
	}
	if err = writeLongstr(w, msg.Locales); err != nil {
		return
	}

	return
}

func (msg *ConnectionStart) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.VersionMajor); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.VersionMinor); err != nil {
		return
	}

	if msg.ServerProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanisms, err = readLongstr(r); err != nil {
		return
	}
	if msg.Locales, err = readLongstr(r); err != nil {
		return
	}

	return
}

type ConnectionStartOk struct {
	ClientProperties Table
	Mechanism        string
	Response         string
	Locale           string
}

func (msg *ConnectionStartOk) id() (uint16, uint16) {
	return 10, 11
}

func (msg *ConnectionStartOk) wait() bool {
	return true
}

func (msg *ConnectionStartOk) write(w io.Writer) (err error) {

	if err = writeTable(w, msg.ClientProperties); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Mechanism); err != nil {
		return
	}

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Locale); err != nil {
		return
	}

	return
}

func (msg *ConnectionStartOk) read(r io.Reader) (err error) {

	if msg.ClientProperties, err = readTable(r); err != nil {
		return
	}

	if msg.Mechanism, err = readShortstr(r); err != nil {
		return
	}

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	if msg.Locale, err = readShortstr(r); err != nil {
		return
	}

	return
}

type connectionSecure struct {
	Challenge string
}

func (msg *connectionSecure) id() (uint16, uint16) {
	return 10, 20
}

func (msg *connectionSecure) wait() bool {
	return true
}

func (msg *connectionSecure) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Challenge); err != nil {
		return
	}

	return
}

func (msg *connectionSecure) read(r io.Reader) (err error) {

	if msg.Challenge, err = readLongstr(r); err != nil {
		return
	}

	return
}

type connectionSecureOk struct {
	Response string
}

func (msg *connectionSecureOk) id() (uint16, uint16) {
	return 10, 21
}

func (msg *connectionSecureOk) wait() bool {
	return true
}

func (msg *connectionSecureOk) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.Response); err != nil {
		return
	}

	return
}

func (msg *connectionSecureOk) read(r io.Reader) (err error) {

	if msg.Response, err = readLongstr(r); err != nil {
		return
	}

	return
}

type connectionTune struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

func (msg *connectionTune) id() (uint16, uint16) {
	return 10, 30
}

func (msg *connectionTune) wait() bool {
	return true
}

func (msg *connectionTune) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

func (msg *connectionTune) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

type connectionTuneOk struct {
	ChannelMax uint16
	FrameMax   uint32
	Heartbeat  uint16
}

func (msg *connectionTuneOk) id() (uint16, uint16) {
	return 10, 31
}

func (msg *connectionTuneOk) wait() bool {
	return true
}

func (msg *connectionTuneOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.FrameMax); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.Heartbeat); err != nil {
		return
	}

	return
}

func (msg *connectionTuneOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ChannelMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.FrameMax); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.Heartbeat); err != nil {
		return
	}

	return
}

type connectionOpen struct {
	VirtualHost string
	reserved1   string
	reserved2   bool
}

func (msg *connectionOpen) id() (uint16, uint16) {
	return 10, 40
}

func (msg *connectionOpen) wait() bool {
	return true
}

func (msg *connectionOpen) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.VirtualHost); err != nil {
		return
	}
	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	if msg.reserved2 {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *connectionOpen) read(r io.Reader) (err error) {
	var bits byte

	if msg.VirtualHost, err = readShortstr(r); err != nil {
		return
	}
	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.reserved2 = (bits&(1<<0) > 0)

	return
}

type connectionOpenOk struct {
	reserved1 string
}

func (msg *connectionOpenOk) id() (uint16, uint16) {
	return 10, 41
}

func (msg *connectionOpenOk) wait() bool {
	return true
}

func (msg *connectionOpenOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *connectionOpenOk) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type ConnectionClose struct {
	ReplyCode uint16 `json:"relyCode"`
	ReplyText string `json:"replyText"`
	ClassId   uint16 `json:"classId"`
	MethodId  uint16 `json:"methodId"`
}

func (msg *ConnectionClose) id() (uint16, uint16) {
	return 10, 50
}

func (msg *ConnectionClose) wait() bool {
	return true
}

func (msg *ConnectionClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

func (msg *ConnectionClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

type ConnectionCloseOk struct {
}

func (msg *ConnectionCloseOk) id() (uint16, uint16) {
	return 10, 51
}

func (msg *ConnectionCloseOk) wait() bool {
	return true
}

func (msg *ConnectionCloseOk) write(w io.Writer) (err error) {

	return
}

func (msg *ConnectionCloseOk) read(r io.Reader) (err error) {

	return
}

type connectionBlocked struct {
	Reason string
}

func (msg *connectionBlocked) id() (uint16, uint16) {
	return 10, 60
}

func (msg *connectionBlocked) wait() bool {
	return false
}

func (msg *connectionBlocked) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Reason); err != nil {
		return
	}

	return
}

func (msg *connectionBlocked) read(r io.Reader) (err error) {

	if msg.Reason, err = readShortstr(r); err != nil {
		return
	}

	return
}

type connectionUnblocked struct {
}

func (msg *connectionUnblocked) id() (uint16, uint16) {
	return 10, 61
}

func (msg *connectionUnblocked) wait() bool {
	return false
}

func (msg *connectionUnblocked) write(w io.Writer) (err error) {

	return
}

func (msg *connectionUnblocked) read(r io.Reader) (err error) {

	return
}

type channelOpen struct {
	reserved1 string
}

func (msg *channelOpen) id() (uint16, uint16) {
	return 20, 10
}

func (msg *channelOpen) wait() bool {
	return true
}

func (msg *channelOpen) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *channelOpen) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type channelOpenOk struct {
	reserved1 string
}

func (msg *channelOpenOk) id() (uint16, uint16) {
	return 20, 11
}

func (msg *channelOpenOk) wait() bool {
	return true
}

func (msg *channelOpenOk) write(w io.Writer) (err error) {

	if err = writeLongstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *channelOpenOk) read(r io.Reader) (err error) {

	if msg.reserved1, err = readLongstr(r); err != nil {
		return
	}

	return
}

type channelFlow struct {
	Active bool
}

func (msg *channelFlow) id() (uint16, uint16) {
	return 20, 20
}

func (msg *channelFlow) wait() bool {
	return true
}

func (msg *channelFlow) write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *channelFlow) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

type channelFlowOk struct {
	Active bool
}

func (msg *channelFlowOk) id() (uint16, uint16) {
	return 20, 21
}

func (msg *channelFlowOk) wait() bool {
	return false
}

func (msg *channelFlowOk) write(w io.Writer) (err error) {
	var bits byte

	if msg.Active {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *channelFlowOk) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Active = (bits&(1<<0) > 0)

	return
}

type channelClose struct {
	ReplyCode uint16
	ReplyText string
	ClassId   uint16
	MethodId  uint16
}

func (msg *channelClose) id() (uint16, uint16) {
	return 20, 40
}

func (msg *channelClose) wait() bool {
	return true
}

func (msg *channelClose) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.ClassId); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.MethodId); err != nil {
		return
	}

	return
}

func (msg *channelClose) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.ClassId); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.MethodId); err != nil {
		return
	}

	return
}

type channelCloseOk struct {
}

func (msg *channelCloseOk) id() (uint16, uint16) {
	return 20, 41
}

func (msg *channelCloseOk) wait() bool {
	return true
}

func (msg *channelCloseOk) write(w io.Writer) (err error) {

	return
}

func (msg *channelCloseOk) read(r io.Reader) (err error) {

	return
}

type ExchangeDeclare struct {
	reserved1  uint16
	Exchange   string `json:"exchange"`
	Type       string `json:"type"`
	Passive    bool   `json:"passive"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"autoDelete"`
	Internal   bool   `json:"internal"`
	NoWait     bool   `json:"noWait"`
	Arguments  Table  `json:"arguments"`
}

func (msg *ExchangeDeclare) id() (uint16, uint16) {
	return 40, 10
}

func (msg *ExchangeDeclare) wait() bool {
	return true && !msg.NoWait
}

func (msg *ExchangeDeclare) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Type); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.AutoDelete {
		bits |= 1 << 2
	}

	if msg.Internal {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *ExchangeDeclare) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.Type, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.AutoDelete = (bits&(1<<2) > 0)
	msg.Internal = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type ExchangeDeclareOk struct {
}

func (msg *ExchangeDeclareOk) id() (uint16, uint16) {
	return 40, 11
}

func (msg *ExchangeDeclareOk) wait() bool {
	return true
}

func (msg *ExchangeDeclareOk) write(w io.Writer) (err error) {

	return
}

func (msg *ExchangeDeclareOk) read(r io.Reader) (err error) {

	return
}

type exchangeDelete struct {
	reserved1 uint16
	Exchange  string
	IfUnused  bool
	NoWait    bool
}

func (msg *exchangeDelete) id() (uint16, uint16) {
	return 40, 20
}

func (msg *exchangeDelete) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeDelete) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.NoWait {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *exchangeDelete) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.NoWait = (bits&(1<<1) > 0)

	return
}

type exchangeDeleteOk struct {
}

func (msg *exchangeDeleteOk) id() (uint16, uint16) {
	return 40, 21
}

func (msg *exchangeDeleteOk) wait() bool {
	return true
}

func (msg *exchangeDeleteOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeDeleteOk) read(r io.Reader) (err error) {

	return
}

type exchangeBind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

func (msg *exchangeBind) id() (uint16, uint16) {
	return 40, 30
}

func (msg *exchangeBind) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeBind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *exchangeBind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type exchangeBindOk struct {
}

func (msg *exchangeBindOk) id() (uint16, uint16) {
	return 40, 31
}

func (msg *exchangeBindOk) wait() bool {
	return true
}

func (msg *exchangeBindOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeBindOk) read(r io.Reader) (err error) {

	return
}

type exchangeUnbind struct {
	reserved1   uint16
	Destination string
	Source      string
	RoutingKey  string
	NoWait      bool
	Arguments   Table
}

func (msg *exchangeUnbind) id() (uint16, uint16) {
	return 40, 40
}

func (msg *exchangeUnbind) wait() bool {
	return true && !msg.NoWait
}

func (msg *exchangeUnbind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Destination); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Source); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *exchangeUnbind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Destination, err = readShortstr(r); err != nil {
		return
	}
	if msg.Source, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type exchangeUnbindOk struct {
}

func (msg *exchangeUnbindOk) id() (uint16, uint16) {
	return 40, 51
}

func (msg *exchangeUnbindOk) wait() bool {
	return true
}

func (msg *exchangeUnbindOk) write(w io.Writer) (err error) {

	return
}

func (msg *exchangeUnbindOk) read(r io.Reader) (err error) {

	return
}

type QueueDeclare struct {
	reserved1  uint16
	Queue      string `json:"queue"`
	Passive    bool   `json:"passive"`
	Durable    bool   `json:"durable"`
	Exclusive  bool   `json:"exclusive"`
	AutoDelete bool   `json:"autoDelete"`
	NoWait     bool   `json:"noWait"`
	Arguments  Table  `json:"arguments"`
}

func (msg *QueueDeclare) id() (uint16, uint16) {
	return 50, 10
}

func (msg *QueueDeclare) wait() bool {
	return true && !msg.NoWait
}

func (msg *QueueDeclare) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.Passive {
		bits |= 1 << 0
	}

	if msg.Durable {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.AutoDelete {
		bits |= 1 << 3
	}

	if msg.NoWait {
		bits |= 1 << 4
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *QueueDeclare) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Passive = (bits&(1<<0) > 0)
	msg.Durable = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.AutoDelete = (bits&(1<<3) > 0)
	msg.NoWait = (bits&(1<<4) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type QueueDeclareOk struct {
	Queue         string
	MessageCount  uint32
	ConsumerCount uint32
}

func (msg *QueueDeclareOk) id() (uint16, uint16) {
	return 50, 11
}

func (msg *QueueDeclareOk) wait() bool {
	return true
}

func (msg *QueueDeclareOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}
	if err = binary.Write(w, binary.BigEndian, msg.ConsumerCount); err != nil {
		return
	}

	return
}

func (msg *QueueDeclareOk) read(r io.Reader) (err error) {

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}
	if err = binary.Read(r, binary.BigEndian, &msg.ConsumerCount); err != nil {
		return
	}

	return
}

type QueueBind struct {
	reserved1  uint16
	Queue      string `json:"queue"`
	Exchange   string `json:"exchange"`
	RoutingKey string `json:"routingKey"`
	NoWait     bool   `json:"noWait"`
	Arguments  Table  `json:"arguments"`
}

func (msg *QueueBind) id() (uint16, uint16) {
	return 50, 20
}

func (msg *QueueBind) wait() bool {
	return true && !msg.NoWait
}

func (msg *QueueBind) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *QueueBind) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type QueueBindOk struct {
}

func (msg *QueueBindOk) id() (uint16, uint16) {
	return 50, 21
}

func (msg *QueueBindOk) wait() bool {
	return true
}

func (msg *QueueBindOk) write(w io.Writer) (err error) {

	return
}

func (msg *QueueBindOk) read(r io.Reader) (err error) {

	return
}

type queueUnbind struct {
	reserved1  uint16
	Queue      string
	Exchange   string
	RoutingKey string
	Arguments  Table
}

func (msg *queueUnbind) id() (uint16, uint16) {
	return 50, 50
}

func (msg *queueUnbind) wait() bool {
	return true
}

func (msg *queueUnbind) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *queueUnbind) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type queueUnbindOk struct {
}

func (msg *queueUnbindOk) id() (uint16, uint16) {
	return 50, 51
}

func (msg *queueUnbindOk) wait() bool {
	return true
}

func (msg *queueUnbindOk) write(w io.Writer) (err error) {

	return
}

func (msg *queueUnbindOk) read(r io.Reader) (err error) {

	return
}

type queuePurge struct {
	reserved1 uint16
	Queue     string
	NoWait    bool
}

func (msg *queuePurge) id() (uint16, uint16) {
	return 50, 30
}

func (msg *queuePurge) wait() bool {
	return true && !msg.NoWait
}

func (msg *queuePurge) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *queuePurge) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

type queuePurgeOk struct {
	MessageCount uint32
}

func (msg *queuePurgeOk) id() (uint16, uint16) {
	return 50, 31
}

func (msg *queuePurgeOk) wait() bool {
	return true
}

func (msg *queuePurgeOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *queuePurgeOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type queueDelete struct {
	reserved1 uint16
	Queue     string
	IfUnused  bool
	IfEmpty   bool
	NoWait    bool
}

func (msg *queueDelete) id() (uint16, uint16) {
	return 50, 40
}

func (msg *queueDelete) wait() bool {
	return true && !msg.NoWait
}

func (msg *queueDelete) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.IfUnused {
		bits |= 1 << 0
	}

	if msg.IfEmpty {
		bits |= 1 << 1
	}

	if msg.NoWait {
		bits |= 1 << 2
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *queueDelete) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.IfUnused = (bits&(1<<0) > 0)
	msg.IfEmpty = (bits&(1<<1) > 0)
	msg.NoWait = (bits&(1<<2) > 0)

	return
}

type queueDeleteOk struct {
	MessageCount uint32
}

func (msg *queueDeleteOk) id() (uint16, uint16) {
	return 50, 41
}

func (msg *queueDeleteOk) wait() bool {
	return true
}

func (msg *queueDeleteOk) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *queueDeleteOk) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type basicQos struct {
	PrefetchSize  uint32
	PrefetchCount uint16
	Global        bool
}

func (msg *basicQos) id() (uint16, uint16) {
	return 60, 10
}

func (msg *basicQos) wait() bool {
	return true
}

func (msg *basicQos) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.PrefetchCount); err != nil {
		return
	}

	if msg.Global {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicQos) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchSize); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.PrefetchCount); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Global = (bits&(1<<0) > 0)

	return
}

type basicQosOk struct {
}

func (msg *basicQosOk) id() (uint16, uint16) {
	return 60, 11
}

func (msg *basicQosOk) wait() bool {
	return true
}

func (msg *basicQosOk) write(w io.Writer) (err error) {

	return
}

func (msg *basicQosOk) read(r io.Reader) (err error) {

	return
}

type BasicConsume struct {
	reserved1   uint16
	Queue       string `json:"queue"`
	ConsumerTag string `json:"consumerTag"`
	NoLocal     bool   `json:"noLocal"`
	NoAck       bool   `json:"noAck"`
	Exclusive   bool   `json:"exclusive"`
	NoWait      bool   `json:"noWait"`
	Arguments   Table  `json:"arguments"`
}

func (msg *BasicConsume) id() (uint16, uint16) {
	return 60, 20
}

func (msg *BasicConsume) wait() bool {
	return true && !msg.NoWait
}

func (msg *BasicConsume) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}
	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoLocal {
		bits |= 1 << 0
	}

	if msg.NoAck {
		bits |= 1 << 1
	}

	if msg.Exclusive {
		bits |= 1 << 2
	}

	if msg.NoWait {
		bits |= 1 << 3
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeTable(w, msg.Arguments); err != nil {
		return
	}

	return
}

func (msg *BasicConsume) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}
	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoLocal = (bits&(1<<0) > 0)
	msg.NoAck = (bits&(1<<1) > 0)
	msg.Exclusive = (bits&(1<<2) > 0)
	msg.NoWait = (bits&(1<<3) > 0)

	if msg.Arguments, err = readTable(r); err != nil {
		return
	}

	return
}

type BasicConsumeOk struct {
	ConsumerTag string
}

func (msg *BasicConsumeOk) id() (uint16, uint16) {
	return 60, 21
}

func (msg *BasicConsumeOk) wait() bool {
	return true
}

func (msg *BasicConsumeOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

func (msg *BasicConsumeOk) read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicCancel struct {
	ConsumerTag string
	NoWait      bool
}

func (msg *basicCancel) id() (uint16, uint16) {
	return 60, 30
}

func (msg *basicCancel) wait() bool {
	return true && !msg.NoWait
}

func (msg *basicCancel) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if msg.NoWait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicCancel) read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoWait = (bits&(1<<0) > 0)

	return
}

type basicCancelOk struct {
	ConsumerTag string
}

func (msg *basicCancelOk) id() (uint16, uint16) {
	return 60, 31
}

func (msg *basicCancelOk) wait() bool {
	return true
}

func (msg *basicCancelOk) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	return
}

func (msg *basicCancelOk) read(r io.Reader) (err error) {

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	return
}

type BasicPublish struct {
	reserved1  uint16
	Exchange   string     `json:"exchange"`
	RoutingKey string     `json:"routingKey"`
	Mandatory  bool       `json:"mandatory"`
	Immediate  bool       `json:"immediate"`
	Properties Properties `json:"properties"`
	Body       []byte     `json:"body"`
}

func (msg *BasicPublish) id() (uint16, uint16) {
	return 60, 40
}

func (msg *BasicPublish) wait() bool {
	return false
}

func (msg *BasicPublish) getContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *BasicPublish) setContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *BasicPublish) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if msg.Mandatory {
		bits |= 1 << 0
	}

	if msg.Immediate {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *BasicPublish) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Mandatory = (bits&(1<<0) > 0)
	msg.Immediate = (bits&(1<<1) > 0)

	return
}

type basicReturn struct {
	ReplyCode  uint16
	ReplyText  string
	Exchange   string
	RoutingKey string
	Properties Properties
	Body       []byte
}

func (msg *basicReturn) id() (uint16, uint16) {
	return 60, 50
}

func (msg *basicReturn) wait() bool {
	return false
}

func (msg *basicReturn) getContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicReturn) setContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicReturn) write(w io.Writer) (err error) {

	if err = binary.Write(w, binary.BigEndian, msg.ReplyCode); err != nil {
		return
	}

	if err = writeShortstr(w, msg.ReplyText); err != nil {
		return
	}
	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

func (msg *basicReturn) read(r io.Reader) (err error) {

	if err = binary.Read(r, binary.BigEndian, &msg.ReplyCode); err != nil {
		return
	}

	if msg.ReplyText, err = readShortstr(r); err != nil {
		return
	}
	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

type BasicDeliver struct {
	ConsumerTag string     `json:"consumerTag"`
	DeliveryTag uint64     `json:"deliveryTag"`
	Redelivered bool       `json:"redelivered"`
	Exchange    string     `json:"exchange"`
	RoutingKey  string     `json:"routingKey"`
	Properties  Properties `json:"properties"`
	Body        []byte     `json:"body"`
}

func (msg *BasicDeliver) id() (uint16, uint16) {
	return 60, 60
}

func (msg *BasicDeliver) wait() bool {
	return false
}

func (msg *BasicDeliver) getContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *BasicDeliver) setContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *BasicDeliver) write(w io.Writer) (err error) {
	var bits byte

	if err = writeShortstr(w, msg.ConsumerTag); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	return
}

func (msg *BasicDeliver) read(r io.Reader) (err error) {
	var bits byte

	if msg.ConsumerTag, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicGet struct {
	reserved1 uint16
	Queue     string
	NoAck     bool
}

func (msg *basicGet) id() (uint16, uint16) {
	return 60, 70
}

func (msg *basicGet) wait() bool {
	return true
}

func (msg *basicGet) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.reserved1); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Queue); err != nil {
		return
	}

	if msg.NoAck {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicGet) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.reserved1); err != nil {
		return
	}

	if msg.Queue, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.NoAck = (bits&(1<<0) > 0)

	return
}

type basicGetOk struct {
	DeliveryTag  uint64
	Redelivered  bool
	Exchange     string
	RoutingKey   string
	MessageCount uint32
	Properties   Properties
	Body         []byte
}

func (msg *basicGetOk) id() (uint16, uint16) {
	return 60, 71
}

func (msg *basicGetOk) wait() bool {
	return true
}

func (msg *basicGetOk) getContent() (Properties, []byte) {
	return msg.Properties, msg.Body
}

func (msg *basicGetOk) setContent(props Properties, body []byte) {
	msg.Properties, msg.Body = props, body
}

func (msg *basicGetOk) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Redelivered {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	if err = writeShortstr(w, msg.Exchange); err != nil {
		return
	}
	if err = writeShortstr(w, msg.RoutingKey); err != nil {
		return
	}

	if err = binary.Write(w, binary.BigEndian, msg.MessageCount); err != nil {
		return
	}

	return
}

func (msg *basicGetOk) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Redelivered = (bits&(1<<0) > 0)

	if msg.Exchange, err = readShortstr(r); err != nil {
		return
	}
	if msg.RoutingKey, err = readShortstr(r); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &msg.MessageCount); err != nil {
		return
	}

	return
}

type basicGetEmpty struct {
	reserved1 string
}

func (msg *basicGetEmpty) id() (uint16, uint16) {
	return 60, 72
}

func (msg *basicGetEmpty) wait() bool {
	return true
}

func (msg *basicGetEmpty) write(w io.Writer) (err error) {

	if err = writeShortstr(w, msg.reserved1); err != nil {
		return
	}

	return
}

func (msg *basicGetEmpty) read(r io.Reader) (err error) {

	if msg.reserved1, err = readShortstr(r); err != nil {
		return
	}

	return
}

type basicAck struct {
	DeliveryTag uint64
	Multiple    bool
}

func (msg *basicAck) id() (uint16, uint16) {
	return 60, 80
}

func (msg *basicAck) wait() bool {
	return false
}

func (msg *basicAck) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicAck) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)

	return
}

type basicReject struct {
	DeliveryTag uint64
	Requeue     bool
}

func (msg *basicReject) id() (uint16, uint16) {
	return 60, 90
}

func (msg *basicReject) wait() bool {
	return false
}

func (msg *basicReject) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicReject) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecoverAsync struct {
	Requeue bool
}

func (msg *basicRecoverAsync) id() (uint16, uint16) {
	return 60, 100
}

func (msg *basicRecoverAsync) wait() bool {
	return false
}

func (msg *basicRecoverAsync) write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicRecoverAsync) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecover struct {
	Requeue bool
}

func (msg *basicRecover) id() (uint16, uint16) {
	return 60, 110
}

func (msg *basicRecover) wait() bool {
	return true
}

func (msg *basicRecover) write(w io.Writer) (err error) {
	var bits byte

	if msg.Requeue {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicRecover) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Requeue = (bits&(1<<0) > 0)

	return
}

type basicRecoverOk struct {
}

func (msg *basicRecoverOk) id() (uint16, uint16) {
	return 60, 111
}

func (msg *basicRecoverOk) wait() bool {
	return true
}

func (msg *basicRecoverOk) write(w io.Writer) (err error) {

	return
}

func (msg *basicRecoverOk) read(r io.Reader) (err error) {

	return
}

type basicNack struct {
	DeliveryTag uint64
	Multiple    bool
	Requeue     bool
}

func (msg *basicNack) id() (uint16, uint16) {
	return 60, 120
}

func (msg *basicNack) wait() bool {
	return false
}

func (msg *basicNack) write(w io.Writer) (err error) {
	var bits byte

	if err = binary.Write(w, binary.BigEndian, msg.DeliveryTag); err != nil {
		return
	}

	if msg.Multiple {
		bits |= 1 << 0
	}

	if msg.Requeue {
		bits |= 1 << 1
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *basicNack) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &msg.DeliveryTag); err != nil {
		return
	}

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Multiple = (bits&(1<<0) > 0)
	msg.Requeue = (bits&(1<<1) > 0)

	return
}

type txSelect struct {
}

func (msg *txSelect) id() (uint16, uint16) {
	return 90, 10
}

func (msg *txSelect) wait() bool {
	return true
}

func (msg *txSelect) write(w io.Writer) (err error) {

	return
}

func (msg *txSelect) read(r io.Reader) (err error) {

	return
}

type txSelectOk struct {
}

func (msg *txSelectOk) id() (uint16, uint16) {
	return 90, 11
}

func (msg *txSelectOk) wait() bool {
	return true
}

func (msg *txSelectOk) write(w io.Writer) (err error) {

	return
}

func (msg *txSelectOk) read(r io.Reader) (err error) {

	return
}

type txCommit struct {
}

func (msg *txCommit) id() (uint16, uint16) {
	return 90, 20
}

func (msg *txCommit) wait() bool {
	return true
}

func (msg *txCommit) write(w io.Writer) (err error) {

	return
}

func (msg *txCommit) read(r io.Reader) (err error) {

	return
}

type txCommitOk struct {
}

func (msg *txCommitOk) id() (uint16, uint16) {
	return 90, 21
}

func (msg *txCommitOk) wait() bool {
	return true
}

func (msg *txCommitOk) write(w io.Writer) (err error) {

	return
}

func (msg *txCommitOk) read(r io.Reader) (err error) {

	return
}

type txRollback struct {
}

func (msg *txRollback) id() (uint16, uint16) {
	return 90, 30
}

func (msg *txRollback) wait() bool {
	return true
}

func (msg *txRollback) write(w io.Writer) (err error) {

	return
}

func (msg *txRollback) read(r io.Reader) (err error) {

	return
}

type txRollbackOk struct {
}

func (msg *txRollbackOk) id() (uint16, uint16) {
	return 90, 31
}

func (msg *txRollbackOk) wait() bool {
	return true
}

func (msg *txRollbackOk) write(w io.Writer) (err error) {

	return
}

func (msg *txRollbackOk) read(r io.Reader) (err error) {

	return
}

type confirmSelect struct {
	Nowait bool
}

func (msg *confirmSelect) id() (uint16, uint16) {
	return 85, 10
}

func (msg *confirmSelect) wait() bool {
	return true
}

func (msg *confirmSelect) write(w io.Writer) (err error) {
	var bits byte

	if msg.Nowait {
		bits |= 1 << 0
	}

	if err = binary.Write(w, binary.BigEndian, bits); err != nil {
		return
	}

	return
}

func (msg *confirmSelect) read(r io.Reader) (err error) {
	var bits byte

	if err = binary.Read(r, binary.BigEndian, &bits); err != nil {
		return
	}
	msg.Nowait = (bits&(1<<0) > 0)

	return
}

type confirmSelectOk struct {
}

func (msg *confirmSelectOk) id() (uint16, uint16) {
	return 85, 11
}

func (msg *confirmSelectOk) wait() bool {
	return true
}

func (msg *confirmSelectOk) write(w io.Writer) (err error) {

	return
}

func (msg *confirmSelectOk) read(r io.Reader) (err error) {

	return
}

func (r *AmqpReader) parseMethodFrame(channel uint16, size uint32) (f frame, err error) {
	mf := &MethodFrame{
		ChannelId: channel,
	}

	if err = binary.Read(r.R, binary.BigEndian, &mf.ClassId); err != nil {
		return
	}

	if err = binary.Read(r.R, binary.BigEndian, &mf.MethodId); err != nil {
		return
	}

	switch mf.ClassId {

	case 10: // connection
		switch mf.MethodId {

		case 10: // connection start
			//fmt.Println("NextMethod: class:10 method:10")
			method := &ConnectionStart{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // connection start-ok
			//fmt.Println("NextMethod: class:10 method:11")
			method := &ConnectionStartOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // connection secure
			//fmt.Println("NextMethod: class:10 method:20")
			method := &connectionSecure{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // connection secure-ok
			//fmt.Println("NextMethod: class:10 method:21")
			method := &connectionSecureOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 30: // connection tune
			//fmt.Println("NextMethod: class:10 method:30")
			method := &connectionTune{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 31: // connection tune-ok
			//fmt.Println("NextMethod: class:10 method:31")
			method := &connectionTuneOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 40: // connection open
			//fmt.Println("NextMethod: class:10 method:40")
			method := &connectionOpen{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 41: // connection open-ok
			//fmt.Println("NextMethod: class:10 method:41")
			method := &connectionOpenOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 50: // connection close
			//fmt.Println("NextMethod: class:10 method:50")
			method := &ConnectionClose{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 51: // connection close-ok
			//fmt.Println("NextMethod: class:10 method:51")
			method := &ConnectionCloseOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 60: // connection blocked
			//fmt.Println("NextMethod: class:10 method:60")
			method := &connectionBlocked{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 61: // connection unblocked
			//fmt.Println("NextMethod: class:10 method:61")
			method := &connectionUnblocked{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 20: // channel
		switch mf.MethodId {

		case 10: // channel open
			//fmt.Println("NextMethod: class:20 method:10")
			method := &channelOpen{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // channel open-ok
			//fmt.Println("NextMethod: class:20 method:11")
			method := &channelOpenOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // channel flow
			//fmt.Println("NextMethod: class:20 method:20")
			method := &channelFlow{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // channel flow-ok
			//fmt.Println("NextMethod: class:20 method:21")
			method := &channelFlowOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 40: // channel close
			//fmt.Println("NextMethod: class:20 method:40")
			method := &channelClose{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 41: // channel close-ok
			//fmt.Println("NextMethod: class:20 method:41")
			method := &channelCloseOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 40: // exchange
		switch mf.MethodId {

		case 10: // exchange declare
			//fmt.Println("NextMethod: class:40 method:10")
			method := &ExchangeDeclare{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // exchange declare-ok
			//fmt.Println("NextMethod: class:40 method:11")
			method := &ExchangeDeclareOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // exchange delete
			//fmt.Println("NextMethod: class:40 method:20")
			method := &exchangeDelete{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // exchange delete-ok
			//fmt.Println("NextMethod: class:40 method:21")
			method := &exchangeDeleteOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 30: // exchange bind
			//fmt.Println("NextMethod: class:40 method:30")
			method := &exchangeBind{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 31: // exchange bind-ok
			//fmt.Println("NextMethod: class:40 method:31")
			method := &exchangeBindOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 40: // exchange unbind
			//fmt.Println("NextMethod: class:40 method:40")
			method := &exchangeUnbind{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 51: // exchange unbind-ok
			//fmt.Println("NextMethod: class:40 method:51")
			method := &exchangeUnbindOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 50: // queue
		switch mf.MethodId {

		case 10: // queue declare
			//fmt.Println("NextMethod: class:50 method:10")
			method := &QueueDeclare{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // queue declare-ok
			//fmt.Println("NextMethod: class:50 method:11")
			method := &QueueDeclareOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // queue bind
			//fmt.Println("NextMethod: class:50 method:20")
			method := &QueueBind{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // queue bind-ok
			//fmt.Println("NextMethod: class:50 method:21")
			method := &QueueBindOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 50: // queue unbind
			//fmt.Println("NextMethod: class:50 method:50")
			method := &queueUnbind{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 51: // queue unbind-ok
			//fmt.Println("NextMethod: class:50 method:51")
			method := &queueUnbindOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 30: // queue purge
			//fmt.Println("NextMethod: class:50 method:30")
			method := &queuePurge{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 31: // queue purge-ok
			//fmt.Println("NextMethod: class:50 method:31")
			method := &queuePurgeOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 40: // queue delete
			//fmt.Println("NextMethod: class:50 method:40")
			method := &queueDelete{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 41: // queue delete-ok
			//fmt.Println("NextMethod: class:50 method:41")
			method := &queueDeleteOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 60: // basic
		switch mf.MethodId {

		case 10: // basic qos
			//fmt.Println("NextMethod: class:60 method:10")
			method := &basicQos{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // basic qos-ok
			//fmt.Println("NextMethod: class:60 method:11")
			method := &basicQosOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // basic consume
			//fmt.Println("NextMethod: class:60 method:20")
			method := &BasicConsume{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // basic consume-ok
			//fmt.Println("NextMethod: class:60 method:21")
			method := &BasicConsumeOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 30: // basic cancel
			//fmt.Println("NextMethod: class:60 method:30")
			method := &basicCancel{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 31: // basic cancel-ok
			//fmt.Println("NextMethod: class:60 method:31")
			method := &basicCancelOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 40: // basic publish
			//fmt.Println("NextMethod: class:60 method:40")
			method := &BasicPublish{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 50: // basic return
			//fmt.Println("NextMethod: class:60 method:50")
			method := &basicReturn{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 60: // basic deliver
			//fmt.Println("NextMethod: class:60 method:60")
			method := &BasicDeliver{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 70: // basic get
			//fmt.Println("NextMethod: class:60 method:70")
			method := &basicGet{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 71: // basic get-ok
			//fmt.Println("NextMethod: class:60 method:71")
			method := &basicGetOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 72: // basic get-empty
			//fmt.Println("NextMethod: class:60 method:72")
			method := &basicGetEmpty{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 80: // basic ack
			//fmt.Println("NextMethod: class:60 method:80")
			method := &basicAck{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 90: // basic reject
			//fmt.Println("NextMethod: class:60 method:90")
			method := &basicReject{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 100: // basic recover-async
			//fmt.Println("NextMethod: class:60 method:100")
			method := &basicRecoverAsync{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 110: // basic recover
			//fmt.Println("NextMethod: class:60 method:110")
			method := &basicRecover{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 111: // basic recover-ok
			//fmt.Println("NextMethod: class:60 method:111")
			method := &basicRecoverOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 120: // basic nack
			//fmt.Println("NextMethod: class:60 method:120")
			method := &basicNack{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 90: // tx
		switch mf.MethodId {

		case 10: // tx select
			//fmt.Println("NextMethod: class:90 method:10")
			method := &txSelect{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // tx select-ok
			//fmt.Println("NextMethod: class:90 method:11")
			method := &txSelectOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 20: // tx commit
			//fmt.Println("NextMethod: class:90 method:20")
			method := &txCommit{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 21: // tx commit-ok
			//fmt.Println("NextMethod: class:90 method:21")
			method := &txCommitOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 30: // tx rollback
			//fmt.Println("NextMethod: class:90 method:30")
			method := &txRollback{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 31: // tx rollback-ok
			//fmt.Println("NextMethod: class:90 method:31")
			method := &txRollbackOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	case 85: // confirm
		switch mf.MethodId {

		case 10: // confirm select
			//fmt.Println("NextMethod: class:85 method:10")
			method := &confirmSelect{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		case 11: // confirm select-ok
			//fmt.Println("NextMethod: class:85 method:11")
			method := &confirmSelectOk{}
			if err = method.read(r.R); err != nil {
				return
			}
			mf.Method = method

		default:
			return nil, ErrBadMethodFrameUnknownMethod
		}

	default:
		return nil, ErrBadMethodFrameUnknownClass
	}

	return mf, nil
}
