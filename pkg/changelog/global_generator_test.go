package changelog

import "testing"

func TestGetSemverStrWeight(t *testing.T) {
	tests := []struct {
		name    string
		verStr  string
		want    int
		wantErr bool
	}{
		{
			name:    "3.4.1",
			verStr:  "3.4.1",
			want:    341,
			wantErr: false,
		},
		{
			name:    "3.4",
			verStr:  "3.4",
			want:    34,
			wantErr: false,
		},
		{
			name:    "3.4-x",
			verStr:  "3.4-x",
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSemverStrWeight(tt.verStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSemverStrWeight() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetSemverStrWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}
