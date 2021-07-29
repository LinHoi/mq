package interceptor

import (
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func DefaultErrorToCode(err error) codes.Code {
	err = errors.Cause(err)
	if err == gorm.ErrRecordNotFound || err == redis.Nil  {
		return codes.NotFound
	}

	return status.Code(err)
}
