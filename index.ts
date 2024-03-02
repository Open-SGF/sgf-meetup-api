import {
	IResource,
	LambdaIntegration,
	MockIntegration,
	PassthroughBehavior,
	RestApi,
} from 'aws-cdk-lib/aws-apigateway';
import { AttributeType, Table } from 'aws-cdk-lib/aws-dynamodb';
import { Runtime } from 'aws-cdk-lib/aws-lambda';
import { App, Stack, RemovalPolicy, Duration } from 'aws-cdk-lib';
import {
	NodejsFunction,
	NodejsFunctionProps,
} from 'aws-cdk-lib/aws-lambda-nodejs';
import { Rule, Schedule } from 'aws-cdk-lib/aws-events';
import { LambdaFunction } from 'aws-cdk-lib/aws-events-targets';
import { Secret } from 'aws-cdk-lib/aws-secretsmanager';
import { join } from 'path';

export class ApiLambdaCrudDynamoDBStack extends Stack {
	constructor(app: App, id: string) {
		super(app, id);

		const NODE_ENV = process.env.BUILD_ENV ?? 'development';

		const EVENTS_TABLE_NAME = 'Events';
		const EVENTS_ID_INDEX_NAME = 'EventsById';
		const EVENTS_GROUP_INDEX_NAME = 'EventsByGroupIndex';

		const eventsTable = new Table(this, EVENTS_TABLE_NAME, {
			partitionKey: {
				name: 'Id',
				type: AttributeType.STRING,
			},
			tableName: EVENTS_TABLE_NAME,
			removalPolicy: RemovalPolicy.RETAIN,
		});

		eventsTable.addGlobalSecondaryIndex({
			indexName: EVENTS_GROUP_INDEX_NAME,
			partitionKey: {
				name: 'MeetupGroupUrlName',
				type: AttributeType.STRING,
			},
			sortKey: {
				name: 'EventDateTime',
				type: AttributeType.STRING,
			},
		});

		const IMPORTER_LOG_TABLE_NAME = 'ImporterLog';

		const importerLogTable = new Table(this, IMPORTER_LOG_TABLE_NAME, {
			partitionKey: {
				name: 'Id',
				type: AttributeType.STRING,
			},
			sortKey: {
				name: 'StartedAt',
				type: AttributeType.STRING,
			},
			tableName: IMPORTER_LOG_TABLE_NAME,
			removalPolicy: RemovalPolicy.DESTROY,
		});

		const API_KEYS = process.env.API_KEYS!;
		const MEETUP_GROUP_NAMES = process.env.MEETUP_GROUP_NAMES!;

		const nodeJsFunctionProps: NodejsFunctionProps = {
			depsLockFilePath: join(__dirname, 'lambdas', 'package-lock.json'),
			environment: {
				LAMBDA_AWS_ACCESS_KEY_ID: 'anything',
				LAMBDA_AWS_SECRET_ACCESS_KEY: 'at-all',
				NODE_ENV,
				EVENTS_TABLE_NAME,
				IMPORTER_LOG_TABLE_NAME,
				EVENTS_GROUP_INDEX_NAME,
				EVENTS_ID_INDEX_NAME,
				API_KEYS,
				MEETUP_GROUP_NAMES,
			},
			runtime: Runtime.NODEJS_18_X,
			timeout: Duration.minutes(4),
			bundling: {
				commandHooks: {
					beforeBundling(
						inputDir: string,
						outputDir: string,
					): string[] {
						return [
							`cp ${inputDir}/.env ${outputDir} 2>/dev/null || :`,
						];
					},
					beforeInstall(): string[] {
						return [];
					},
					afterBundling(): string[] {
						return [];
					},
				},
			},
		};

		const importerLambda = new NodejsFunction(this, 'importerFunction', {
			entry: join(__dirname, 'lambdas', 'importer.ts'),
			...nodeJsFunctionProps,
		});

		const meetupKeySecret = Secret.fromSecretAttributes(
			this,
			'MeetupKeySecret',
			{
				secretCompleteArn:
					'arn:aws:secretsmanager:us-east-2:391849688676:secret:prod/sgf-meetup-api/meetup-UbNhVU',
			},
		);

		meetupKeySecret.grantRead(importerLambda);

		const getEventsLambda = new NodejsFunction(this, 'getEventsFunction', {
			entry: join(__dirname, 'lambdas', 'events.ts'),
			...nodeJsFunctionProps,
		});

		// Grant the Lambda function read access to the DynamoDB table
		// getEventsLambda.node.addDependency(eventsTable);
		// getEventsLambda.permissionsNode.addDependency(eventsTable);
		// importerLambda.node.addDependency(importerLogTable);
		// importerLambda.permissionsNode.addDependency(importerLogTable);
		eventsTable.grantReadWriteData(getEventsLambda);
		importerLogTable.grantReadWriteData(importerLambda);

		const importScheduleRule = new Rule(this, 'importerEventBridgeRule', {
			// schedule: Schedule.expression('cron(0 0-23/2 * * ? *)'), // "run every 2 hours"
			schedule: Schedule.expression('cron(0-59/2 * * * ? *)'), // "run every 2 minutes"
		});

		importScheduleRule.addTarget(new LambdaFunction(importerLambda));

		// Integrate the Lambda functions with the API Gateway resource
		const getEventsIntegration = new LambdaIntegration(getEventsLambda);

		// Create an API Gateway resource for each of the CRUD operations
		const api = new RestApi(this, 'eventsApi', {
			restApiName: 'Events Service',
			// In case you want to manage binary types, uncomment the following
			// binaryMediaTypes: ["*/*"],
		});

		const eventsResource = api.root.addResource('events');
		eventsResource.addMethod('GET', getEventsIntegration);
		addCorsOptions(eventsResource);
	}
}

export function addCorsOptions(apiResource: IResource) {
	apiResource.addMethod(
		'OPTIONS',
		new MockIntegration({
			// In case you want to use binary media types, uncomment the following line
			// contentHandling: ContentHandling.CONVERT_TO_TEXT,
			integrationResponses: [
				{
					statusCode: '200',
					responseParameters: {
						'method.response.header.Access-Control-Allow-Headers':
							"'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
						'method.response.header.Access-Control-Allow-Origin':
							"'*'",
						'method.response.header.Access-Control-Allow-Credentials':
							"'false'",
						'method.response.header.Access-Control-Allow-Methods':
							"'OPTIONS,GET,PUT,POST,DELETE'",
					},
				},
			],
			// In case you want to use binary media types, comment out the following line
			passthroughBehavior: PassthroughBehavior.NEVER,
			requestTemplates: {
				'application/json': '{"statusCode": 200}',
			},
		}),
		{
			methodResponses: [
				{
					statusCode: '200',
					responseParameters: {
						'method.response.header.Access-Control-Allow-Headers':
							true,
						'method.response.header.Access-Control-Allow-Methods':
							true,
						'method.response.header.Access-Control-Allow-Credentials':
							true,
						'method.response.header.Access-Control-Allow-Origin':
							true,
					},
				},
			],
		},
	);
}

const app = new App();
new ApiLambdaCrudDynamoDBStack(app, 'sgf-meetup-api');
app.synth();
