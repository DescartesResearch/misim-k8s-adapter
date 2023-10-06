package control

import (
	"go-kube/pkg/storage"
	"strconv"
)

type IdGenerator struct {
	idStorage storage.IdStorage
}

func (s *IdGenerator) GetNextResourceId() string {
	current := s.idStorage.GetNextId()
	result := strconv.Itoa(current)
	next := current + 1
	s.idStorage.StoreNextId(next)
	return result
}
