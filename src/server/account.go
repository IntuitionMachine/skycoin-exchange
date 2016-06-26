package skycoin-exchange

import (
	"fmt"
	"sync"

	"github.com/skycoin/skycoin-exchange/src/server/wallet"
	"github.com/skycoin/skycoin/src/cipher"
)

type AccountID cipher.PubKey
type Balance uint64 // satoshis

type Account interface {
	GetNewAddress(ct wallet.CoinType) string // return new address for receiveing coins
	GetBalance(ct wallet.CoinType) string    // return the current balance.
	// GetUnspentOutput(ct wallet.CoinType, minConf int) // return all unspent output of this account that confirms minConf times.
}

// AccountState maintains the account state
type AccountState struct {
	ID          AccountID                   // account id
	balance     map[wallet.CoinType]Balance // the balance should not be accessed directly.
	wltID       string                      // wallet used to maintain the address, UTXOs, balance, etc.
	balance_mtx sync.RWMutex                // mutex used to protect the balance's concurrent read and write.
	wlt_mtx     sync.Mutex                  // mutex used to protect the wallet's conncurrent read and write.
}

func newAccountState(id AccountID, wltID string) AccountState {
	return AccountState{
		ID:    id,
		wltID: wltID,
		balance: map[wallet.CoinType]Balance{
			wallet.Bitcoin: 0,
			wallet.Skycoin: 0,
		}}
}

// GetNewAddress generate new address for this account.
func (self *AccountState) GetNewAddress(ct wallet.CoinType) string {
	// get the wallet.
	wlt, err := wallet.GetWallet(self.wltID)
	if err != nil {
		panic(fmt.Sprintf("account get wallet faild, wallet id:%s", self.wltID))
	}

	self.wlt_mtx.Lock()
	defer self.wlt_mtx.Unlock()
	addr, err := wlt.NewAddresses(ct, 1)
	if err != nil {
		panic(err)
	}
	return addr[0].Address
}

// Get the current recored balance.
func (self *AccountState) GetBalance(coinType wallet.CoinType) (Balance, error) {
	self.balance_mtx.RLock()
	defer self.balance_mtx.RUnlock()
	if b, ok := self.balance[coinType]; ok {
		return b, nil
	}
	return 0, fmt.Errorf("the account does not have %s", coinType)
}

// SetBalance update the balanace of specific coin.
func (self *AccountState) setBalance(coinType wallet.CoinType, balance Balance) error {
	self.balance_mtx.Lock()
	defer self.balance_mtx.Unlock()
	if _, ok := self.balance[coinType]; !ok {
		return fmt.Errorf("the account does not have %s", coinType)
	}
	self.balance[coinType] = balance
	return nil
}