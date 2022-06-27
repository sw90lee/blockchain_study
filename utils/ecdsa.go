package utils

import (
	"fmt"
	"math/big"
)

type Signature struct {
	// 공개키 X 좌표
	R *big.Int
	// Transation Hash와 임시같은 정보를 참조하여 계산 가능
	S *big.Int
}

func (s *Signature) String() string {
	return fmt.Sprintf("%x%x", s.R, s.S)
}
