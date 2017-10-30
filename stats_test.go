package main

import "testing"

func TestCanGetStats(t *testing.T) {
	w := getWalletAddress()
	s, err := GetStats(w)
	if err != nil {
		t.Error(err)
	}
	if s.Result.Addr != w {
		t.Error("Address not same.")
	}
	b := s.GetBalance()
	if b <= 0 {
		t.Error("Balance is zero.")
	}
}
