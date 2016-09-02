package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/skycoin/skycoin-exchange/src/client/account"
	"github.com/skycoin/skycoin-exchange/src/coin"
	bitcoin "github.com/skycoin/skycoin-exchange/src/coin/bitcoin"
	skycoin "github.com/skycoin/skycoin-exchange/src/coin/skycoin"
	"github.com/skycoin/skycoin-exchange/src/pp"
	"github.com/skycoin/skycoin-exchange/src/sknet"
	"github.com/skycoin/skycoin/src/cipher"
)

// InjectTx broadcast transaction.
func InjectTx(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var rlt *pp.EmptyRes
		for {
			a, err := account.GetActive()
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}
			// get tx
			tx := r.FormValue("tx")
			if tx == "" {
				logger.Error("empty tx")
				rlt = pp.MakeErrRes(errors.New("empty tx"))
				break
			}

			// get coin type
			cp := r.FormValue("coin_type")
			if cp == "" {
				logger.Error("empty coin type")
				rlt = pp.MakeErrRes(errors.New("empty coin type"))
				break
			}

			req := pp.InjectTxnReq{
				CoinType: pp.PtrString(cp),
				Tx:       pp.PtrString(tx),
			}

			encReq, err := makeEncryptReq(&req, se.GetServKey().Hex(), a.Seckey)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_WrongRequest)
				break
			}

			resp, err := sknet.Get(se.GetServAddr(), "/auth/inject/tx", encReq)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			v, err := decodeRsp(resp.Body, se.GetServKey().Hex(), a.Seckey, &pp.InjectTxnRes{})
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			sendJSON(w, v)
			return
		}
		sendJSON(w, rlt)
	}
}

// GetTx get transaction.
func GetTx(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var rlt *pp.EmptyRes
		for {
			a, err := account.GetActive()
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}
			// get coin type
			cp := r.FormValue("coin_type")
			if cp == "" {
				rlt = pp.MakeErrRes(errors.New("no coin type"))
				break
			}

			// get txid
			txid := r.FormValue("txid")
			if txid == "" {
				rlt = pp.MakeErrRes(errors.New("no txid"))
				break
			}
			req := pp.GetTxReq{
				CoinType: pp.PtrString(cp),
				Txid:     pp.PtrString(txid),
			}
			encReq, err := makeEncryptReq(req, se.GetServKey().Hex(), a.Seckey)
			if err != nil {
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			rsp, err := sknet.Get(se.GetServAddr(), "/auth/get/tx", encReq)
			if err != nil {
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			res, err := decodeRsp(rsp.Body, se.GetServKey().Hex(), a.Seckey, &pp.GetTxRes{})
			if err != nil {
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			sendJSON(w, res)
			return
		}
		logger.Error(rlt.GetResult().GetReason())
		sendJSON(w, rlt)
	}
}

// GetRawTx get raw tx from exchange server.
func GetRawTx(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var rlt *pp.EmptyRes
		for {
			a, err := account.GetActive()
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}
			// get coin type
			cp := r.FormValue("coin_type")
			if cp == "" {
				logger.Error("no coin type")
				rlt = pp.MakeErrRes(errors.New("no coin type"))
				break
			}
			// get txid
			txid := r.FormValue("txid")
			if txid == "" {
				logger.Error("no txid")
				rlt = pp.MakeErrRes(errors.New("no txid"))
				break
			}
			req := pp.GetRawTxReq{
				CoinType: pp.PtrString(cp),
				Txid:     pp.PtrString(txid),
			}
			encReq, err := makeEncryptReq(req, se.GetServKey().Hex(), a.Seckey)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			rsp, err := sknet.Get(se.GetServAddr(), "/auth/get/rawtx", encReq)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			res, err := decodeRsp(rsp.Body, se.GetServKey().Hex(), a.Seckey, &pp.GetRawTxRes{})
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			sendJSON(w, res)
			return
		}
		sendJSON(w, rlt)
	}
}

type rawTxParams struct {
	TxIns  []coin.TxIn `json:"tx_ins"`
	TxOuts []struct {
		Addr  string `json:"address"`
		Value uint64 `json:"value"`
		Hours uint64 `json:"hours"`
	} `json:"tx_outs"`
}

// CreateRawTx create raw tx base on some utxos.
// mode: POST
// url: /api/v1/rawtx?coin_type=[:coin_type]
// request body:
func CreateRawTx(se Servicer) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var rlt *pp.EmptyRes
	loop:
		for {
			// get coin type
			cp, err := coin.TypeFromStr(r.FormValue("coin_type"))
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}

			// get request body
			params := rawTxParams{}
			if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}

			gw, err := coin.GetGateway(cp)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			var rawtx string
			switch cp {
			case coin.Bitcoin:
				outs := make([]bitcoin.TxOut, len(params.TxOuts))
				for i, o := range params.TxOuts {
					outs[i].Addr = o.Addr
					outs[i].Value = o.Value
				}
				rawtx, err = gw.CreateRawTx(params.TxIns, outs)
			case coin.Skycoin:
				outs := make([]skycoin.TxOut, len(params.TxOuts))
				for i, o := range params.TxOuts {
					addr, err := cipher.DecodeBase58Address(o.Addr)
					if err != nil {
						logger.Error(err.Error())
						rlt = pp.MakeErrRes(err)
						break loop
					}
					outs[i].Address = addr
					outs[i].Coins = o.Value
					outs[i].Hours = o.Hours
				}
				rawtx, err = gw.CreateRawTx(params.TxIns, outs)
			}
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrRes(err)
				break
			}

			res := struct {
				Result *pp.Result `json:"result"`
				Rawtx  string     `json:"rawtx"`
			}{
				Result: pp.MakeResultWithCode(pp.ErrCode_Success),
				Rawtx:  rawtx,
			}
			sendJSON(w, &res)
			return
		}
		sendJSON(w, rlt)
	}
}
