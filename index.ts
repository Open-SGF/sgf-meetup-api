import { join } from 'path';
import {
	BasePathMapping,
	DomainName,
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
import { Policy, Role } from 'aws-cdk-lib/aws-iam';

const AWS_ACCOUNT_ID = '391849688676';
const AWS_REGION = 'us-east-2';
const MEETUP_KEY_ARN =
	'arn:aws:secretsmanager:us-east-2:391849688676:secret:prod/sgf-meetup-api/meetup-UbNhVU';
const SECRETS_ACCESS_ROLE_ARN =
	'arn:aws:iam::391849688676:role/aws-reserved/sso.amazonaws.com/us-east-2/AWSReservedSSO_SgfMeetupApiSecretOnlyAccess_3c2a00e4cc366092';

const NODE_ENV = process.env.BUILD_ENV ?? 'development';
const EVENTS_TABLE_NAME = 'Events';
const IMPORTER_LOG_TABLE_NAME = 'ImporterLog';
const EVENTS_ID_INDEX_NAME = 'EventsById';
const EVENTS_GROUP_INDEX_NAME = 'EventsByGroupIndex';
const ROOT_DOMAIN = 'opensgf.org';
const EVENTS_API_SUBDOMAIN = 'sgf-meetup-api';
const EVENTS_API_DOMAIN_NAME = `${EVENTS_API_SUBDOMAIN}.${ROOT_DOMAIN}`;
const GET_MEETUP_TOKEN_FUNCTION_NAME = 'getMeetupTokenFunction';

// user/client info
const API_KEYS = process.env.API_KEYS!;
const MEETUP_GROUP_NAMES = process.env.MEETUP_GROUP_NAMES!;

export class ApiLambdaCrudDynamoDBStack extends Stack {
	constructor(app: App, id: string) {
		super(app, id, {
			env: {
				account: AWS_ACCOUNT_ID,
				region: AWS_REGION,
			},
		});

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
			removalPolicy: RemovalPolicy.RETAIN,
		});

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
				GET_MEETUP_TOKEN_FUNCTION_NAME,
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
			'meetupKeySecret',
			{ secretCompleteArn: MEETUP_KEY_ARN },
		);

		const getEventsLambda = new NodejsFunction(this, 'getEventsFunction', {
			entry: join(__dirname, 'lambdas', 'events.ts'),
			...nodeJsFunctionProps,
		});

		const getMeetupTokenLambda = new NodejsFunction(
			this,
			GET_MEETUP_TOKEN_FUNCTION_NAME,
			{
				entry: join(__dirname, 'lambdas', 'getMeetupToken.ts'),
				functionName: GET_MEETUP_TOKEN_FUNCTION_NAME,
				...nodeJsFunctionProps,
			},
		);

		new Policy(this, 'getMeetupTokenLambdaInvokePolicy', {
            policyName: 'getMeetupTokenLambdaInvokePolicy',
            statements: [
                new iam.PolicyStatement({
                    actions: ['lambda:InvokeFunction'],
                    resources: [getMeetupTokenLambda.functionArn],
                    effect: iam.Effect.ALLOW
                })
            ]
        });

		const getMeetupTokenRole = Role.fromRoleArn(
			this,
			'getMeetupTokenRole',
			SECRETS_ACCESS_ROLE_ARN,
		);

		meetupKeySecret.grantRead(getMeetupTokenLambda);
		getMeetupTokenLambda.grantInvoke(importerLambda);
		getMeetupTokenLambda.grantInvoke(getMeetupTokenRole);
		eventsTable.grantReadWriteData(getEventsLambda);
		eventsTable.grantReadWriteData(importerLambda);
		importerLogTable.grantReadWriteData(importerLambda);

		const importScheduleRule = new Rule(this, 'importerEventBridgeRule', {
			schedule: Schedule.expression('cron(0 0-23/2 * * ? *)'), // "run every 2 hours"
			// schedule: Schedule.expression('cron(0-59/2 * * * ? *)'), // "run every 2 minutes"
		});

		importScheduleRule.addTarget(new LambdaFunction(importerLambda));

		// Integrate the Lambda functions with the API Gateway resource
		const getEventsIntegration = new LambdaIntegration(getEventsLambda);

		// const certificate = acm.Certificate.fromCertificateArn(
		// 	this,
		// 	'domainCert',
		// 	'arn:aws:acm:us-east-2:391849688676:certificate/c64e30b4-1531-4357-bf80-672b4d8978c8',
		// );

		const domainName = DomainName.fromDomainNameAttributes(
			this,
			'eventsApiDomainName',
			{
				domainName: EVENTS_API_DOMAIN_NAME,
				domainNameAliasHostedZoneId: 'ZOJJZC49E0EPZ',
				domainNameAliasTarget:
					'd-x4jexiktj7.execute-api.us-east-2.amazonaws.com',
			},
		);

		// Create an API Gateway resource for each of the CRUD operations
		const restApi = new RestApi(this, 'eventsApi', {
			restApiName: 'Events Service',
			// In case you want to manage binary types, uncomment the following
			// binaryMediaTypes: ["*/*"],
		});

		new BasePathMapping(this, 'apiBasePathMapping', {
			domainName,
			restApi,
		});

		const eventsResource = restApi.root.addResource('events');
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
