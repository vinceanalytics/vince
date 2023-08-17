package v1

func (s *StoreKey) Parts() []string {
	if s.Namespace == "" {
		s.Namespace = "vince"
	}
	return []string{
		s.Namespace, s.Prefix.String(),
	}
}

func (s *Site_Key) Parts() []string {
	return append(s.Store.Parts(), s.Domain)
}

func (s *Block_Key) Parts() []string {
	return append(s.Store.Parts(), s.Kind.String(), s.Domain, s.Uid)
}

func (s *Account_Key) Parts() []string {
	return append(s.Store.Parts(), s.Name)
}
