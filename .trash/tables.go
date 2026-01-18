package cli

func InitTables(tables []string) map[string]string {
	out := make(map[string]string, 0)
	for _, v := range tables {
		out[v] = v
	}
	return out
}
