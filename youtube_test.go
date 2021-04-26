package main

import (
	"testing"
)

func TestGetChatcolor(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testGetChatcolor(t, 100, 0x134A9D)
		testGetChatcolor(t, 199, 0x134A9D)
		testGetChatcolor(t, 200, 0x28E4FD)
		testGetChatcolor(t, 499, 0x28E4FD)
		testGetChatcolor(t, 500, 0x32E8B7)
		testGetChatcolor(t, 999, 0x32E8B7)
		testGetChatcolor(t, 1000, 0xFCD748)
		testGetChatcolor(t, 1999, 0xFCD748)
		testGetChatcolor(t, 2000, 0xF37C22)
		testGetChatcolor(t, 4999, 0xF37C22)
		testGetChatcolor(t, 5000, 0xE72564)
		testGetChatcolor(t, 9999, 0xE72564)
		testGetChatcolor(t, 10000, 0xE32624)
		testGetChatcolor(t, 50000, 0xE32624)
	})
	t.Run("failed", func(t *testing.T) {
		_, err := getChatcolor(99)
		if (err == nil) {
			t.Errorf("less than 100 yen, don't raise error'")
		}
	})
}

func testGetChatcolor(t *testing.T, pay, expected int) {
	t.Helper()

	color, err := getChatcolor(pay)
	if err != nil {
		t.Fatal(err)
	}
	if color != expected {
		t.Errorf("Compute(%d) = %d, want %d", pay, color, expected)
	}
}
