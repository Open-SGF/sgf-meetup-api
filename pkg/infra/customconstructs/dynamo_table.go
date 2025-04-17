package customconstructs

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
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
	tableName := *tableProps.TableName
	fullTableName := tableName

	if prefix != "" {
		fullTableName = prefix + tableName
		tableProps.TableName = jsii.String(fullTableName)
	}

	table := awsdynamodb.NewTable(scope, jsii.String(tableName), &tableProps)

	for _, gsi := range props.GlobalSecondaryIndexes {
		table.AddGlobalSecondaryIndex(&gsi)
	}

	return &DynamoTable{
		Table:         table,
		FullTableName: fullTableName,
	}
}
