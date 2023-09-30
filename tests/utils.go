package tests

import "testing"

func CheckError(t *testing.T, wantErr, gotErr error) {
	t.Helper()

	if wantErr == nil && gotErr != nil {
		t.Errorf("got error %v, want nil", gotErr)

		return
	}

	if wantErr != nil && gotErr == nil {
		t.Errorf("got nil, want error %v", wantErr)

		return
	}

	if wantErr != nil && gotErr != nil && gotErr.Error() != wantErr.Error() {
		t.Errorf("got error %v, want %v", gotErr, wantErr)

		return
	}
}
