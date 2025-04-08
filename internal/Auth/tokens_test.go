package auth

import (
	"fmt"
	"testing"

	config "github.com/hegner123/modulacms/internal/config"
)

func TestGenerateAccessToken(t *testing.T) {
	verbose := false
	c := config.LoadConfig(&verbose, "")
	s, err := GenerateAccessToken(1, 1, c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)

}

func TestGenerateRefreshToken(t *testing.T) {
	verbose := false
	c := config.LoadConfig(&verbose, "")
	s, err := GenerateRefreshToken(1, 1, c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(s)

}

func TestReadToken(t *testing.T) {
	verbose := false
	c := config.LoadConfig(&verbose, "")
	s, err := GenerateRefreshToken(1, 1, c)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := ReadToken(s, c)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range claims {
		fmt.Printf(" %v:%v\n", k, v)
	}

}
