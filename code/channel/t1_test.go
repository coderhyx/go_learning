package channel

import "testing"

func TestReadCloseCh1(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "测试",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("读已关闭通道")
				}
			}()
			ReadCloseCh1()
		})
	}
}
