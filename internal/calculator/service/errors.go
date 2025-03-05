package service

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func InternalError(err error) error {
	return status.Error(codes.Internal, fmt.Sprintf("oops, something went wrong: %s", err.Error()))
}
