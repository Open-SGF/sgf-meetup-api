import 'dotenv/config';
import * as fs from "fs";
import { ScanCommand } from "@aws-sdk/client-dynamodb";
import { dynamoDBClient } from "./dynamodb-client";

async function getAll() {
  const tableName = "items";
  console.log({ tableName });
  const command = new ScanCommand({ TableName: tableName });

  const response = await dynamoDBClient.send(command);
  console.log({ response });
  return response.Items;
}

export const handler = async (): Promise<any> => {
  const result = await getAll();
  return { statusCode: 200, body: JSON.stringify({ hello: 'VERLD', result }) };
};
