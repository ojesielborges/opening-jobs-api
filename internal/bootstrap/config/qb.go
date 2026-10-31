package config

import "github.com/Masterminds/squirrel"

func InitializeQueryBuilder() *squirrel.StatementBuilderType {
	queryBuilder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	return &queryBuilder
}
