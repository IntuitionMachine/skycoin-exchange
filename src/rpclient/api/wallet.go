package api

import (
	"errors"
	"net/http"

	"github.com/skycoin/skycoin-exchange/src/coin"
	"github.com/skycoin/skycoin-exchange/src/pp"
	"github.com/skycoin/skycoin-exchange/src/wallet"
)

// CreateWallet api for creating local wallet.
// mode: POST
// url: /api/v1/wallet?coin_type=[:coin_type]&seed=[:seed]
// params:
// 		coin_type: bitcoin or skycoin
// 		seed: wallet seed.
func CreateWallet(se Servicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rlt := &pp.EmptyRes{}
		for {
			// check method
			if r.Method != "POST" {
				rlt = pp.MakeErrRes(errors.New("require POST method"))
				break
			}

			// get coin type
			cp, err := coin.TypeFromStr(r.FormValue("coin_type"))
			if err != nil {
				rlt = pp.MakeErrRes(err)
				break
			}

			// get seed
			sd := r.FormValue("seed")
			if sd == "" {
				rlt = pp.MakeErrRes(errors.New("no seed"))
				break
			}

			wlt, err := wallet.New(cp, sd)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			res := struct {
				Result *pp.Result `json:"result"`
				ID     string     `json:"id"`
			}{
				Result: pp.MakeResultWithCode(pp.ErrCode_Success),
				ID:     wlt.GetID(),
			}
			sendJSON(w, &res)
			return
		}
		sendJSON(w, rlt)
	}
}

// NewAddress create address in wallet.
// mode: POST
// url: /api/v1/wallet/address?&wallet_id=[:wallet_id]
// params:
// 		wallet_id: wallet id.
func NewAddress(se Servicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rlt *pp.EmptyRes
		for {
			if r.Method != "POST" {
				rlt = pp.MakeErrRes(errors.New("require POST method"))
				break
			}

			// get wallet id
			wltID := r.FormValue("wallet_id")
			if wltID == "" {
				rlt = pp.MakeErrRes(errors.New("no wallet_id"))
				break
			}

			addrEntries, err := wallet.NewAddresses(wltID, 1)
			if err != nil {
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}
			res := struct {
				Result  *pp.Result
				Address string `json:"address"`
			}{
				Result:  pp.MakeResultWithCode(pp.ErrCode_Success),
				Address: addrEntries[0].Address,
			}
			sendJSON(w, &res)
			return
		}
		sendJSON(w, rlt)
	}
}

// GetKeys get keys of specific address in wallet.
// mode: GET
// url: /api/v1/wallet/address/keys?wallet_id=[:wallet_id]&address=[:address]
func GetKeys(se Servicer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rlt *pp.EmptyRes
		for {
			if r.Method != "GET" {
				rlt = pp.MakeErrRes(errors.New("require GET method"))
				break
			}

			// get wallet id
			wltID := r.FormValue("wallet_id")
			if wltID == "" {
				rlt = pp.MakeErrRes(errors.New("no wallet id"))
				break
			}

			// get address
			addr := r.FormValue("address")
			if addr == "" {
				rlt = pp.MakeErrRes(errors.New("no address"))
				break
			}
			p, s, err := wallet.GetKeypair(wltID, addr)
			if err != nil {
				logger.Error(err.Error())
				rlt = pp.MakeErrResWithCode(pp.ErrCode_ServerError)
				break
			}

			res := struct {
				Result *pp.Result
				Pubkey string `json:"pubkey"`
				Seckey string `json:"seckey"`
			}{
				Result: pp.MakeResultWithCode(pp.ErrCode_Success),
				Pubkey: p,
				Seckey: s,
			}
			sendJSON(w, &res)
			return
		}
		sendJSON(w, rlt)
	}
}
