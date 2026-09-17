package db

import "testing"

func TestBucketLikePattern(t *testing.T) {
	tests := []struct {
		bucket string
		want   string
	}{
		{bucket: "outbound", want: "outbound:%"},
		{bucket: `foo\%_bar`, want: `foo\\\%\_bar:%`},
	}

	for _, test := range tests {
		if got := bucketLikePattern(test.bucket); got != test.want {
			t.Errorf("bucketLikePattern(%q) = %q, want %q", test.bucket, got, test.want)
		}
	}
}
