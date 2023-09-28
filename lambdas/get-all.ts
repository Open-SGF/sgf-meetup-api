import 'dotenv/config';
import * as fs from "fs";
import fetch from "node-fetch";

import { BatchGetItemCommand, ListTablesCommand, ScanCommand } from "@aws-sdk/client-dynamodb";
import { dynamoDBClient } from "./dynamodb-client";

async function getAllTables() {
  const command = new ListTablesCommand({});
  const response = await dynamoDBClient.send(command);
  return response.TableNames;
}

async function getAll() {
  // const tableName = process.env.TABLE_NAME!; // TODO: understand why this gets passed through as items07D08F4B
  // const primaryKey = process.env.PRIMARY_KEY!;

  const tableName = "items";
  console.log({ tableName });

  // const RequestItems = { [tableName]: { Keys: [ ] }};

  // const params = { RequestItems };

  // const command = new BatchGetItemCommand(params);

  const command = new ScanCommand({ TableName: tableName });

  const response = await dynamoDBClient.send(command);
  return response.Items;
}

export const handler = async (): Promise<any> => {
  // TODO: fix problem with environment variables not getting populated?
  // console.log(".env file contains");
  // const dotenvfile = fs.readFileSync(".env").toString();
  // console.log(dotenvfile);
  console.log(`process.env.SOMETHING_NEW=${process.env.SOMETHING_NEW}`);
  console.log(`process.env.PRIMARY_KEY=${process.env.PRIMARY_KEY}`);
  const result = await getAll();
  return { statusCode: 200, body: JSON.stringify({hello: 'VERLD', result }) };
};
