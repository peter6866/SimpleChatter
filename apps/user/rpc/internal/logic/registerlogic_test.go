package logic

import (
	"context"
	"testing"

	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"
)

func TestRegisterLogic_Register(t *testing.T) {
	type args struct {
		in *user.RegisterReq
	}
	tests := []struct {
		name      string
		args      args
		wantPrint bool
		wantErr   bool
	}{
		{
			"1", args{
				in: &user.RegisterReq{
					Phone:    "12345678902",
					Nickname: "test",
					Avatar:   "test.jpg",
					Password: "123456",
					Sex:      1,
				}}, true, false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewRegisterLogic(context.Background(), svcCtx)
			got, err := l.Register(tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterLogic.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantPrint {
				t.Log(tt.name, got)
			}
		})
	}
}
