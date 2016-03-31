package azure

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenResponse := `{"access_token": "PXVZK7eTrxt4fjpK5s"}`
		fmt.Fprintln(w, tokenResponse)
	}))
	defer ts.Close()

	token, err := getToken(ts.URL, "64443DDC-5263-4AD1-90CD-C3F25F02D56A", "VXZj6VUvTLoGbarbqLkegQUTeFdrIed39E2==")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%s", token)

	if want, got := "PXVZK7eTrxt4fjpK5s", token; want != got.Bearer {
		t.Errorf("want token %s, got %s", want, got.Bearer)
	}
}
