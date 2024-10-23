module github.com/navivan123/gator

go 1.23.2

replace internal/config => ./internal/config/
replace internal/database => ./internal/database/

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	internal/config v1.0.0
	internal/database v1.0.0
)
