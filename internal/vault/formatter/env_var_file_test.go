package formatter

import "testing"

func TestEnvVarFormatterSortsKeys(t *testing.T) {
	data := map[string]string{"user": "app", "password": "s3cret", "host": "db", "port": "5432"}
	want := "HOST=db\nPASSWORD=s3cret\nPORT=5432\nUSER=app\n"
	for range 20 {
		got, err := NewEnvVarFormatter(true, false).Format(data)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("Format() = %q, want %q", got, want)
		}
	}
}
