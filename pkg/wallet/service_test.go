package wallet

import (
	"github.com/sidalsoft/wallet/pkg/types"
	"testing"
)

func TestService_FindAccountByID(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
	}
	_, _ = srv.RegisterAccount("992928563355")
	_, _ = srv.RegisterAccount("992928563352")
	_, _ = srv.RegisterAccount("992928125354")

	_, err := srv.FindAccountByID(2)
	if err != nil {
		if err != ErrAccountNotFound {
			t.Errorf("Invalid result, expected : %v, actual %v", err, ErrAccountNotFound)
		}
	}
}
