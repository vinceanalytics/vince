package v1

import "path"

func (s *StoreKey) Badger() string {
	if s.Namespace == "" {
		s.Namespace = "vince"
	}
	return path.Join(s.Namespace, s.Prefix.String(), s.Key)
}

func (s *Block_Key) Badger() string {
	return (&StoreKey{
		Prefix: StorePrefix_BLOCKS,
		Key:    path.Join(s.Kind.String(), s.Domain, s.Uid),
	}).Badger()
}
