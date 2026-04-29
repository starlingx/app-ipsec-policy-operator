/*
Copyright (c) 2026 Wind River Systems, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"testing"
)

func TestGetNonEmptyLines_Basic(t *testing.T) {
	result := GetNonEmptyLines("line1\nline2\nline3")
	if len(result) != 3 {
		t.Errorf("expected 3 lines, got %d", len(result))
	}
}

func TestGetNonEmptyLines_WithEmpty(t *testing.T) {
	result := GetNonEmptyLines("line1\n\nline2\n\n")
	if len(result) != 2 {
		t.Errorf("expected 2 lines, got %d: %v", len(result), result)
	}
}

func TestGetNonEmptyLines_AllEmpty(t *testing.T) {
	result := GetNonEmptyLines("\n\n\n")
	if len(result) != 0 {
		t.Errorf("expected 0 lines, got %d", len(result))
	}
}

func TestGetNonEmptyLines_EmptyString(t *testing.T) {
	result := GetNonEmptyLines("")
	if len(result) != 0 {
		t.Errorf("expected 0 lines, got %d", len(result))
	}
}

func TestGetNonEmptyLines_SingleLine(t *testing.T) {
	result := GetNonEmptyLines("hello")
	if len(result) != 1 || result[0] != "hello" {
		t.Errorf("expected [hello], got %v", result)
	}
}

func TestGetProjectDir(t *testing.T) {
	dir, err := GetProjectDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty directory")
	}
}
