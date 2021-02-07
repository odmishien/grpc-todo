package interceptor

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/odmishien/grpctodo/config"
	"google.golang.org/grpc"
)

func NewAuthInterceptor() grpc.UnaryServerInterceptor {
	return grpc_auth.UnaryServerInterceptor(firebaseAuth)
}

func firebaseAuth(ctx context.Context) (context.Context, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		panic(err)
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		panic(err)
	}
	claims, err := auth.VerifyIDToken(ctx, token)
	if err != nil {
		return nil, err
	}
	fmt.Printf("user:%s\n", claims.UID)
	ctx = context.WithValue(ctx, config.UserKey, claims.UID)
	return ctx, nil
}
