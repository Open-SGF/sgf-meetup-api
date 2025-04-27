package customconstructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"sgf-meetup-api/pkg/shared/resource"
)

type DynamoTableProps struct {
	awsdynamodb.TableProps
	GlobalSecondaryIndexes []awsdynamodb.GlobalSecondaryIndexProps
}

type DynamoTable struct {
	awsdynamodb.Table
	FullTableName string
}

func NewDynamoTable(scope constructs.Construct, prefix string, props *DynamoTableProps) *DynamoTable {
	tableProps := props.TableProps
	tableNamer := resource.NewNamer(prefix, *tableProps.TableName)

	tableProps.TableName = jsii.String(tableNamer.FullName())

	table := awsdynamodb.NewTable(scope, jsii.String(tableNamer.Name()), &tableProps)

	for _, gsi := range props.GlobalSecondaryIndexes {
		table.AddGlobalSecondaryIndex(&gsi)
	}

	return &DynamoTable{
		Table:         table,
		FullTableName: tableNamer.FullName(),
	}
}
