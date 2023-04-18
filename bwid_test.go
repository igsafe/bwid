package bwid

import (
	"log"
	"strings"
	"testing"
)

func TestGenerateTokenLen(t *testing.T) {
	token := GenerateToken(24)
	log.Printf("GenerateToken(24) %s", token)
	if len(token) != 24 {
		t.Fatalf("Expected length 24, got %d", len(token))
	}
}

func TestGenerateTimestampedTokenLen1(t *testing.T) {
	token := GenerateTimestampedToken(24)
	log.Printf("GenerateTimestampedToken(24) %s", token)
	if len(token) != 24 {
		t.Fatalf("Expected length 24, got %d", len(token))
	}
}

func TestGenerateTimestampedTokenLen2(t *testing.T) {
	token := GenerateTimestampedToken(40)
	log.Printf("GenerateTimestampedToken(40) %s", token)
	if len(token) != 40 {
		t.Fatalf("Expected length 40, got %d", len(token))
	}
}

func TestGenerateObjectIdLen(t *testing.T) {
	token := GenerateObjectId()
	log.Printf("GenerateObjectId() %s", token)
	if len(token) != 24 {
		t.Fatalf("Expected length 24, got %d", len(token))
	}
}

func TestGenerateBulkSeqTimestampedTokenLen(t *testing.T) {
	var count int64 = 100000
	var tokenLen = 40
	tokens := GenerateBulkSeqTimestampedToken(count, 40)
	if int64(len(tokens)) != count {
		t.Fatalf("Expected %d tokens, got %d", count, len(tokens))
	}
	log.Printf("GenerateBulkSeqTimestampedToken(n, 40) %s", tokens[0])
	for _, token := range tokens {
		if len(token) != tokenLen {
			t.Fatalf("Expected token length %d for %s", tokenLen, token)
		}
	}
}

func TestGenerateBulkSeqTimestampedTokenSeq(t *testing.T) {
	tokens := GenerateBulkSeqTimestampedToken(100000, 40)
	var prevToken string
	log.Printf("GenerateBulkSeqTimestampedToken(n, 40) %s", tokens[0])
	for _, token := range tokens {
		if strings.Compare(prevToken, token) != -1 {
			t.Fatalf("Generated out of sequence tokens: '%s' '%s'", prevToken, token)
		}
		prevToken = token
	}
}
