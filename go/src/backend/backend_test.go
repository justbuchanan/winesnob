package main

import (
    "testing"
    "fmt"
)

func TestJoinWordSeries(t *testing.T) {
    items := []string{"red", "blue", "green"}
    result := JoinWordSeries(items)
    expected := "red, blue, and green"
    if result != expected {
        t.Error("result != expected")
    }
}