package main

import (
	"database/sql"
	"time"
)

func ns(s string)sql.NullString{
return sql.NullString{String: s}
}

func ni(i int)sql.NullInt64{
return sql.NullInt64{Int64: int64(i)}
}

func nf(f float64)sql.NullFloat64{
return sql.NullFloat64{Float64: f}
}

func nb(b bool)sql.NullBool{ 
return sql.NullBool{Bool: b}
}

func nt(t time.Time)sql.NullTime{
return sql.NullTime{Time: t}
}

func nby(by byte)sql.NullByte{
return sql.NullByte{Byte: by}
}

