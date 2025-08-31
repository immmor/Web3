package common

import "github.com/zeromicro/go-zero/rest"

func WithCORS() rest.RunOption {
    return rest.WithCors()
}