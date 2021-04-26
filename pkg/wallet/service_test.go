package wallet

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/sidalsoft/wallet/pkg/types"
	"reflect"
	"testing"
)

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	return account, nil
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, nil
}

var defaultTestAccount = testAccount{
	phone:   "+992925556644",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID(): must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}
	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error =%v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed, payment = %v", savedPayment)
		return
	}
	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}

}

func TestService_Repeat_success(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	p, _ := srv.Repeat(pp.ID)
	p.ID = pp.ID
	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Repeat(): expected %v returned = %v", pp, p)
	}
}

func TestService_Repeat_fail(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")

	_, err := srv.Repeat(uuid.New().String())
	if err == nil {
		t.Error("Repeat(): must return error, returned nil")
		return
	}
}

func TestService_FavoritePayment_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, err := srv.FavoritePayment(pp.ID, "sidal")

	if err != nil {
		t.Error("FavoritePayment(): can't make favorite return error, returned nil")
		return
	}
}

func TestService_FavoritePayment_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, err := srv.FavoritePayment(pp.ID, "sidal")
	_, err = srv.FavoritePayment(pp.ID, "sidal")

	if err == nil {
		t.Error("FavoritePayment(): must return error, returned nil")
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	fw, _ := srv.FavoritePayment(pp.ID, "sidal")

	_, err := srv.PayFromFavorite(fw.ID)
	if err != nil {
		t.Error("PayFromFavorite(): can't make favorite return error, returned nil")
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, err := srv.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Error("FavoritePayment(): must return error, returned nil")
		return
	}
}

func TestService_FindFavoriteByID_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	fw, _ := srv.FavoritePayment(pp.ID, "sidal")

	_, err := srv.FindFavoriteByID(fw.ID)
	if err != nil {
		t.Error("FindFavoriteByID(): can't make favorite return error, returned nil")
	}
}

func TestService_FindFavoriteByID_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, err := srv.FindFavoriteByID(uuid.New().String())

	if err == nil {
		t.Error("FindFavoriteByID(): must return error, returned nil")
	}
}

func TestService_ExportToFile(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")
	_, _ = srv.RegisterAccount("+992928000000")
	_, _ = srv.RegisterAccount("+992928811111")
	err := srv.ExportToFile("salom.txt")
	println(err)
}

func TestService_ImportFromFile(t *testing.T) {
	srv := &Service{accounts: make([]*types.Account, 0)}
	err := srv.ImportFromFile("salom.txt")
	println(err)
}

func TestService_Export(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")
	_, _ = srv.RegisterAccount("+992928000000")
	ac, _ := srv.RegisterAccount("+992928811111")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, _ = srv.FavoritePayment(pp.ID, "sidal")
	err := srv.Export("./data")
	exp := srv.accounts
	srv.accounts = append(srv.accounts[0:1], srv.accounts[2:]...)
	//srv.payments = make([]*types.Payment, 0)
	//srv.favorites = make([]*types.Favorite, 0)

	println(exp)
	err = srv.Import("./data")
	err = srv.Export("./data1")
	if err != nil {
		panic(err)
	}
	println(err)
}

func BenchmarkSumPayments(b *testing.B) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	account, err := srv.RegisterAccount("+992926574322")
	if err != nil {
		b.Errorf("account => %v", account)
	}
	err = srv.Deposit(account.ID, 100_00)
	if err != nil {
		b.Errorf("error => %v", err)
	}
	want := types.Money(55)
	for i := types.Money(1); i <= 10; i++ {
		_, err := srv.Pay(account.ID, i, "aa")
		if err != nil {
			b.Errorf("error => %v", err)
		}
	}
	got := srv.SumPayments(5)
	if want != got {
		b.Errorf("want => %v got => %v", want, got)
	}
}
