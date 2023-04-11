package nacosconfig

import (
	"fmt"
	"testing"
)

func TestClient(t *testing.T) {
	c, _ := NewClient("5e49d8ed-c51a-4d59-9b48-4a71b70f65ac", "12344", "11111", "yaml")

	_ = c.SetWatcher()
	fmt.Println(c.GetString("test"))
}
