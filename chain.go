package main

import (
	"github.com/google/uuid"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf"
	"gitlab.db.in.tum.de/c-chain/ccf-go/ccf/apimodels"
	"log"
	"time"
)

/*
This file contains all functions which handle the chains and it's transactions
*/

func mustNot(err error) {
	if err != nil {
		panic(err)
	}
}

func InitTransactionManager() *TAM {
	//session, err := ccf.StartSession("https://localhost", "RSA_WITH_AES_128_CBC", "keys") -> doesn't work with local ccm!
	//session, err := ccf.StartSession("https://ccm.db.in.tum.de:8443", "RSA_WITH_AES_128_CBC", "keys") -> doesn't work (timeout)
	session, err := ccf.StartSession("https://cchaintest.db.in.tum.de:443", "RSA_WITH_AES_128_CBC", "keys")
	if err != nil {
		log.Fatal("Error Init CCF Service", err.Error())
	}
	return &TAM{
		ccf:       session,
		publicKey: session.PublicKeyB,
		signKey:   session.SignedKey,
		expires:   time.Now().Add(time.Hour * 24 * 365),
	}
}

func (T *TAM) CreateChain(name string, desc string, closed bool, publicRead bool, publicWrite bool) (*apimodels.ChainInfo, error) {
	return T.ccf.CreateChain(name, desc, closed, publicRead, publicWrite)
}

func (T *TAM) GetChain(chainID uuid.UUID) (*apimodels.ChainInfo, error) {
	return T.ccf.GetChain(chainID)
}

func (T *TAM) CreateTransaction(chainID uuid.UUID, payload []byte) (*apimodels.Transaction, error) {
	return T.ccf.CreateTransaction(chainID, T.publicKey, T.signKey, payload, 30, false, true)
}

func (T *TAM) GetTransaction(chainID uuid.UUID, transactionID uuid.UUID) (*apimodels.Transaction, error) {
	return T.ccf.GetTransaction(chainID, transactionID)
}

func (T *TAM) Close() error {
	return T.ccf.TerminateSession()
}
