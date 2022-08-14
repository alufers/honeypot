package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/xelabs/go-mysqlstack/driver"
	querypb "github.com/xelabs/go-mysqlstack/sqlparser/depends/query"
	"github.com/xelabs/go-mysqlstack/sqlparser/depends/sqltypes"
	"github.com/xelabs/go-mysqlstack/xlog"
)

type HoneypotMysqlHandler struct {
	driver.TestHandler
	attacks map[uint32]*Attack
}

func NewHoneypotMysqlHandler(logger *xlog.Log) *HoneypotMysqlHandler {
	return &HoneypotMysqlHandler{
		TestHandler: *driver.NewTestHandler(logger),
		attacks:     make(map[uint32]*Attack),
	}
}

func (h *HoneypotMysqlHandler) ComQuery(session *driver.Session, query string, bindVariables map[string]*querypb.BindVariable, callback func(*sqltypes.Result) error) error {

	attack := h.attacks[session.ID()]
	if attack != nil {
		attack.Classification = "command_entered"
		attack.Write([]byte(query + "\n"))
		if len(bindVariables) > 0 {
			attack.Write([]byte(fmt.Sprintf("  bindVariables: %+v\n", bindVariables)))
		}
		AttackUpdated(attack)
	}

	result1 := &sqltypes.Result{
		Fields: []*querypb.Field{
			{
				Name: "id",
				Type: querypb.Type_INT32,
			},
			{
				Name: "name",
				Type: querypb.Type_VARCHAR,
			},
		},
		Rows: [][]sqltypes.Value{
			{
				sqltypes.MakeTrusted(querypb.Type_INT32, []byte("123")),
				sqltypes.MakeTrusted(querypb.Type_VARCHAR, []byte("hi there")),
			},
		},
	}
	return callback(result1)
	// return h.TestHandler.ComQuery(session, query, bindVariables, callback)
}

func (h *HoneypotMysqlHandler) AuthCheck(session *driver.Session) error {
	username := session.User()

	log.Printf("auth check %v, %v", username, session.Addr())
	attack := &Attack{
		Protocol:       "mysql",
		SourceIP:       strings.Split(session.Addr(), ":")[0],
		Contents:       "",
		Action:         "other",
		Classification: "authenticated",
	}
	if err := AttackStarted(attack); err != nil {
		return err
	}
	h.attacks[session.ID()] = attack

	return nil
}

func (h *HoneypotMysqlHandler) SessionClosed(session *driver.Session) {
	attack := h.attacks[session.ID()]
	if attack != nil {
		AttackFinished(attack)
		delete(h.attacks, session.ID())
	}
}

func RunMysqlServer() {
	mysqlPort := 3306
	fmt.Sscanf(getEnv("HONEYPOT_MYSQL_PORT", "3306"), "%d", &mysqlPort)

	log.Printf("Starting mysql server on %v", mysqlPort)
	listeningProtocolsMutex.Lock()
	listeningProtocols = append(listeningProtocols, "mysql")
	listeningProtocolsMutex.Unlock()
	logger := xlog.NewStdLog(xlog.Level(xlog.INFO))
	th := NewHoneypotMysqlHandler(logger)
	mysqld, err := driver.MockMysqlServerWithPort(logger, mysqlPort, th)
	if err != nil {
		log.Panic("mysqld.start.error:%+v", err)
	}
	defer mysqld.Close()
	// block forever
	select {}
}
