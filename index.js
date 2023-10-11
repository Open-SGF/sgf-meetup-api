"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.addCorsOptions = exports.ApiLambdaCrudDynamoDBStack = void 0;
const aws_apigateway_1 = require("aws-cdk-lib/aws-apigateway");
const aws_dynamodb_1 = require("aws-cdk-lib/aws-dynamodb");
const aws_lambda_1 = require("aws-cdk-lib/aws-lambda");
const aws_cdk_lib_1 = require("aws-cdk-lib");
const aws_lambda_nodejs_1 = require("aws-cdk-lib/aws-lambda-nodejs");
const aws_events_1 = require("aws-cdk-lib/aws-events");
const aws_events_targets_1 = require("aws-cdk-lib/aws-events-targets");
const path_1 = require("path");
class ApiLambdaCrudDynamoDBStack extends aws_cdk_lib_1.Stack {
    constructor(app, id) {
        var _a;
        super(app, id);
        const dynamoTable = new aws_dynamodb_1.Table(this, 'items', {
            partitionKey: {
                name: 'itemId',
                type: aws_dynamodb_1.AttributeType.STRING
            },
            tableName: 'items',
            /**
             *  The default removal policy is RETAIN, which means that cdk destroy will not attempt to delete
             * the new table, and it will remain in your account until manually deleted. By setting the policy to
             * DESTROY, cdk destroy will delete the table (even if it has data in it)
             */
            removalPolicy: aws_cdk_lib_1.RemovalPolicy.DESTROY, // NOT recommended for production code
        });
        const NODE_ENV = (_a = process.env.BUILD_ENV) !== null && _a !== void 0 ? _a : "development";
        const nodeJsFunctionProps = {
            depsLockFilePath: (0, path_1.join)(__dirname, 'lambdas', 'package-lock.json'),
            environment: {
                PRIMARY_KEY: 'itemId',
                TABLE_NAME: dynamoTable.tableName,
                LAMBDA_AWS_ACCESS_KEY_ID: "anything",
                LAMBDA_AWS_SECRET_ACCESS_KEY: "at-all",
                NODE_ENV
            },
            runtime: aws_lambda_1.Runtime.NODEJS_18_X,
        };
        const getAllLambda = new aws_lambda_nodejs_1.NodejsFunction(this, 'getAllItemsFunction', {
            entry: (0, path_1.join)(__dirname, 'lambdas', 'get-all.ts'),
            ...nodeJsFunctionProps,
        });
        const importerLambda = new aws_lambda_nodejs_1.NodejsFunction(this, 'importerFunction', {
            entry: (0, path_1.join)(__dirname, 'lambdas', 'importer.ts'),
            bundling: {
                commandHooks: {
                    beforeBundling(inputDir, outputDir) {
                        const commands = [
                            `cp ${inputDir}/meetup-private-key ${outputDir}`,
                        ];
                        if (process.env.BUILD_ENV !== 'production') {
                            commands.push(`cp ${inputDir}/.env ${outputDir}`);
                        }
                        return commands;
                    },
                    beforeInstall() {
                        return [];
                    },
                    afterBundling() {
                        return [];
                    }
                },
            },
            ...nodeJsFunctionProps,
        });
        // Grant the Lambda function read access to the DynamoDB table
        dynamoTable.grantReadWriteData(getAllLambda);
        dynamoTable.grantReadWriteData(importerLambda);
        const importScheduleRule = new aws_events_1.Rule(this, 'importerEventBridgeRule', {
            schedule: aws_events_1.Schedule.expression('cron(0 2 * * ? *)'),
        });
        importScheduleRule.addTarget(new aws_events_targets_1.LambdaFunction(importerLambda));
        // Integrate the Lambda functions with the API Gateway resource
        const getAllIntegration = new aws_apigateway_1.LambdaIntegration(getAllLambda);
        // Create an API Gateway resource for each of the CRUD operations
        const api = new aws_apigateway_1.RestApi(this, 'itemsApi', {
            restApiName: 'Items Service'
            // In case you want to manage binary types, uncomment the following
            // binaryMediaTypes: ["*/*"],
        });
        const items = api.root.addResource('items');
        items.addMethod('GET', getAllIntegration);
        addCorsOptions(items);
    }
}
exports.ApiLambdaCrudDynamoDBStack = ApiLambdaCrudDynamoDBStack;
function addCorsOptions(apiResource) {
    apiResource.addMethod('OPTIONS', new aws_apigateway_1.MockIntegration({
        // In case you want to use binary media types, uncomment the following line
        // contentHandling: ContentHandling.CONVERT_TO_TEXT,
        integrationResponses: [{
                statusCode: '200',
                responseParameters: {
                    'method.response.header.Access-Control-Allow-Headers': "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
                    'method.response.header.Access-Control-Allow-Origin': "'*'",
                    'method.response.header.Access-Control-Allow-Credentials': "'false'",
                    'method.response.header.Access-Control-Allow-Methods': "'OPTIONS,GET,PUT,POST,DELETE'",
                },
            }],
        // In case you want to use binary media types, comment out the following line
        passthroughBehavior: aws_apigateway_1.PassthroughBehavior.NEVER,
        requestTemplates: {
            "application/json": "{\"statusCode\": 200}"
        },
    }), {
        methodResponses: [{
                statusCode: '200',
                responseParameters: {
                    'method.response.header.Access-Control-Allow-Headers': true,
                    'method.response.header.Access-Control-Allow-Methods': true,
                    'method.response.header.Access-Control-Allow-Credentials': true,
                    'method.response.header.Access-Control-Allow-Origin': true,
                },
            }]
    });
}
exports.addCorsOptions = addCorsOptions;
const app = new aws_cdk_lib_1.App();
new ApiLambdaCrudDynamoDBStack(app, 'springfieldMeetupApi');
app.synth();
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSwrREFBeUg7QUFDekgsMkRBQWdFO0FBQ2hFLHVEQUFpRDtBQUNqRCw2Q0FBd0Q7QUFDeEQscUVBQW9GO0FBQ3BGLHVEQUF3RDtBQUN4RCx1RUFBZ0U7QUFDaEUsK0JBQTRCO0FBRTVCLE1BQWEsMEJBQTJCLFNBQVEsbUJBQUs7SUFDbkQsWUFBWSxHQUFRLEVBQUUsRUFBVTs7UUFDOUIsS0FBSyxDQUFDLEdBQUcsRUFBRSxFQUFFLENBQUMsQ0FBQztRQUVmLE1BQU0sV0FBVyxHQUFHLElBQUksb0JBQUssQ0FBQyxJQUFJLEVBQUUsT0FBTyxFQUFFO1lBQzNDLFlBQVksRUFBRTtnQkFDWixJQUFJLEVBQUUsUUFBUTtnQkFDZCxJQUFJLEVBQUUsNEJBQWEsQ0FBQyxNQUFNO2FBQzNCO1lBQ0QsU0FBUyxFQUFFLE9BQU87WUFFbEI7Ozs7ZUFJRztZQUNILGFBQWEsRUFBRSwyQkFBYSxDQUFDLE9BQU8sRUFBRSxzQ0FBc0M7U0FDN0UsQ0FBQyxDQUFDO1FBRUgsTUFBTSxRQUFRLEdBQUcsTUFBQSxPQUFPLENBQUMsR0FBRyxDQUFDLFNBQVMsbUNBQUksYUFBYSxDQUFDO1FBRXhELE1BQU0sbUJBQW1CLEdBQXdCO1lBQy9DLGdCQUFnQixFQUFFLElBQUEsV0FBSSxFQUFDLFNBQVMsRUFBRSxTQUFTLEVBQUUsbUJBQW1CLENBQUM7WUFDakUsV0FBVyxFQUFFO2dCQUNYLFdBQVcsRUFBRSxRQUFRO2dCQUNyQixVQUFVLEVBQUUsV0FBVyxDQUFDLFNBQVM7Z0JBQ2pDLHdCQUF3QixFQUFFLFVBQVU7Z0JBQ3BDLDRCQUE0QixFQUFFLFFBQVE7Z0JBQ3RDLFFBQVE7YUFDVDtZQUNELE9BQU8sRUFBRSxvQkFBTyxDQUFDLFdBQVc7U0FDN0IsQ0FBQTtRQUVELE1BQU0sWUFBWSxHQUFHLElBQUksa0NBQWMsQ0FBQyxJQUFJLEVBQUUscUJB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RUFBRSxJQUFJO29CQUMzRCx5REFBeUQsRUFBRSxJQUFJO29CQUMvRCxvREFBb0QsRUFBRSxJQUFJO2lCQUMzRDthQUNGLENBQUM7S0FDSCxDQUFDLENBQUE7QUFDSixDQUFDO0FBN0JELHdDQTZCQztBQUVELE1BQU0sR0FBRyxHQUFHLElBQUksaUJBQUcsRUFBRSxDQUFDO0FBQ3RCLElBQUksMEJBQTBCLENBQUMsR0FBRyxFQUFFLHNCQUFzQixDQUFDLENBQUM7QUFDNUQsR0FBRyxDQUFDLEtBQUssRUFBRSxDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgSVJlc291cmNlLCBMYW1iZGFJbnRlZ3JhdGlvbiwgTW9ja0ludGVncmF0aW9uLCBQYXNzdGhyb3VnaEJlaGF2aW9yLCBSZXN0QXBpIH0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWFwaWdhdGV3YXknO1xuaW1wb3J0IHsgQXR0cmlidXRlVHlwZSwgVGFibGUgfSBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZHluYW1vZGInO1xuaW1wb3J0IHsgUnVudGltZSB9IGZyb20gJ2F3cy1jZGstbGliL2F3cy1sYW1iZGEnO1xuaW1wb3J0IHsgQXBwLCBTdGFjaywgUmVtb3ZhbFBvbGljeSB9IGZyb20gJ2F3cy1jZGstbGliJztcbmltcG9ydCB7IE5vZGVqc0Z1bmN0aW9uLCBOb2RlanNGdW5jdGlvblByb3BzIH0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWxhbWJkYS1ub2RlanMnO1xuaW1wb3J0IHsgUnVsZSwgU2NoZWR1bGUgfSBmcm9tICdhd3MtY2RrLWxpYi9hd3MtZXZlbnRzJztcbmltcG9ydCB7IExhbWJkYUZ1bmN0aW9uIH0gZnJvbSAnYXdzLWNkay1saWIvYXdzLWV2ZW50cy10YXJnZXRzJztcbmltcG9ydCB7IGpvaW4gfSBmcm9tICdwYXRoJztcblxuZXhwb3J0IGNsYXNzIEFwaUxhbWJkYUNydWREeW5hbW9EQlN0YWNrIGV4dGVuZHMgU3RhY2sge1xuICBjb25zdHJ1Y3RvcihhcHA6IEFwcCwgaWQ6IHN0cmluZykge1xuICAgIHN1cGVyKGFwcCwgaWQpO1xuXG4gICAgY29uc3QgZHluYW1vVGFibGUgPSBuZXcgVGFibGUodGhpcywgJ2l0ZW1zJywge1xuICAgICAgcGFydGl0aW9uS2V5OiB7XG4gICAgICAgIG5hbWU6ICdpdGVtSWQnLFxuICAgICAgICB0eXBlOiBBdHRyaWJ1dGVUeXBlLlNUUklOR1xuICAgICAgfSxcbiAgICAgIHRhYmxlTmFtZTogJ2l0ZW1zJyxcblxuICAgICAgLyoqXG4gICAgICAgKiAgVGhlIGRlZmF1bHQgcmVtb3ZhbCBwb2xpY3kgaXMgUkVUQUlOLCB3aGljaCBtZWFucyB0aGF0IGNkayBkZXN0cm95IHdpbGwgbm90IGF0dGVtcHQgdG8gZGVsZXRlXG4gICAgICAgKiB0aGUgbmV3IHRhYmxlLCBhbmQgaXQgd2lsbCByZW1haW4gaW4geW91ciBhY2NvdW50IHVudGlsIG1hbnVhbGx5IGRlbGV0ZWQuIEJ5IHNldHRpbmcgdGhlIHBvbGljeSB0b1xuICAgICAgICogREVTVFJPWSwgY2RrIGRlc3Ryb3kgd2lsbCBkZWxldGUgdGhlIHRhYmxlIChldmVuIGlmIGl0IGhhcyBkYXRhIGluIGl0KVxuICAgICAgICovXG4gICAgICByZW1vdmFsUG9saWN5OiBSZW1vdmFsUG9saWN5LkRFU1RST1ksIC8vIE5PVCByZWNvbW1lbmRlZCBmb3IgcHJvZHVjdGlvbiBjb2RlXG4gICAgfSk7XG5cbiAgICBjb25zdCBOT0RFX0VOViA9IHByb2Nlc3MuZW52LkJVSUxEX0VOViA/PyBcImRldmVsb3BtZW50XCI7XG5cbiAgICBjb25zdCBub2RlSnNGdW5jdGlvblByb3BzOiBOb2RlanNGdW5jdGlvblByb3BzID0ge1xuICAgICAgZGVwc0xvY2tGaWxlUGF0aDogam9pbihfX2Rpcm5hbWUsICdsYW1iZGFzJywgJ3BhY2thZ2UtbG9jay5qc29uJyksXG4gICAgICBlbnZpcm9ubWVudDoge1xuICAgICAgICBQUklNQVJZX0tFWTogJ2l0ZW1JZCcsXG4gICAgICAgIFRBQkxFX05BTUU6IGR5bmFtb1RhYmxlLnRhYmxlTmFtZSxcbiAgICAgICAgTEFNQkRBX0FXU19BQ0NFU1NfS0VZX0lEOiBcImFueXRoaW5nXCIsXG4gICAgICAgIExBTUJEQV9BV1NfU0VDUkVUX0FDQ0VTU19LRVk6IFwiYXQtYWxsXCIsXG4gICAgICAgIE5PREVfRU5WXG4gICAgICB9LFxuICAgICAgcnVudGltZTogUnVudGltZS5OT0RFSlNfMThfWCxcbiAgICB9XG5cbiAgICBjb25zdCBnZXRBbGxMYW1iZGEgPSBuZXcgTm9kZWpzRnVuY3Rpb24odGhpcywgJ2dldEFsbEl0ZW1zRnVuY3Rpb24nLCB7XG4gICAgICBlbnRyeTogam9pbihfX2Rpcm5hbWUsICdsYW1iZGFzJywgJ2dldC1hbGwudHMnKSxcbiAgICAgIC4uLm5vZGVKc0Z1bmN0aW9uUHJvcHMsXG4gICAgfSk7XG5cbiAgICBjb25zdCBpbXBvcnRlckxhbWJkYSA9IG5ldyBOb2RlanNGdW5jdGlvbih0aGlzLCAnaW1wb3J0ZXJGdW5jdGlvbicsIHtcbiAgICAgIGVudHJ5OiBqb2luKF9fZGlybmFtZSwgJ2xhbWJkYXMnLCAnaW1wb3J0ZXIudHMnKSxcbiAgICAgIGJ1bmRsaW5nOiB7XG4gICAgICAgIGNvbW1hbmRIb29rczoge1xuICAgICAgICAgIGJlZm9yZUJ1bmRsaW5nKGlucHV0RGlyOiBzdHJpbmcsIG91dHB1dERpcjogc3RyaW5nKTogc3RyaW5nW10ge1xuICAgICAgICAgICAgY29uc3QgY29tbWFuZHMgPSBbXG4gICAgICAgICAgICAgIGBjcCAke2lucHV0RGlyfS9tZWV0dXAtcHJpdmF0ZS1rZXkgJHtvdXRwdXREaXJ9YCxcbiAgICAgICAgICAgIF1cblxuICAgICAgICAgICAgaWYgKHByb2Nlc3MuZW52LkJVSUxEX0VOViAhPT0gJ3Byb2R1Y3Rpb24nKSB7XG4gICAgICAgICAgICAgIGNvbW1hbmRzLnB1c2goYGNwICR7aW5wdXREaXJ9Ly5lbnYgJHtvdXRwdXREaXJ9YCwpXG4gICAgICAgICAgICB9XG5cbiAgICAgICAgICAgIHJldHVybiBjb21tYW5kcztcbiAgICAgICAgICB9LFxuICAgICAgICAgIGJlZm9yZUluc3RhbGwoKTogc3RyaW5nW10ge1xuICAgICAgICAgICAgcmV0dXJuIFtdO1xuICAgICAgICAgIH0sXG4gICAgICAgICAgYWZ0ZXJCdW5kbGluZygpOiBzdHJpbmdbXSB7XG4gICAgICAgICAgICByZXR1cm4gW107XG4gICAgICAgICAgfVxuICAgICAgICB9LFxuICAgICAgfSxcbiAgICAgIC4uLm5vZGVKc0Z1bmN0aW9uUHJvcHMsXG4gICAgfSlcblxuICAgIC8vIEdyYW50IHRoZSBMYW1iZGEgZnVuY3Rpb24gcmVhZCBhY2Nlc3MgdG8gdGhlIER5bmFtb0RCIHRhYmxlXG4gICAgZHluYW1vVGFibGUuZ3JhbnRSZWFkV3JpdGVEYXRhKGdldEFsbExhbWJkYSk7XG4gICAgZHluYW1vVGFibGUuZ3JhbnRSZWFkV3JpdGVEYXRhKGltcG9ydGVyTGFtYmRhKTtcblxuICAgIGNvbnN0IGltcG9ydFNjaGVkdWxlUnVsZSA9IG5ldyBSdWxlKHRoaXMsICdpbXBvcnRlckV2ZW50QnJpZGdlUnVsZScsIHtcbiAgICAgIHNjaGVkdWxlOiBTY2hlZHVsZS5leHByZXNzaW9uKCdjcm9uKDAgMiAqICogPyAqKScpLFxuICAgIH0pXG5cbiAgICBpbXBvcnRTY2hlZHVsZVJ1bGUuYWRkVGFyZ2V0KG5ldyBMYW1iZGFGdW5jdGlvbihpbXBvcnRlckxhbWJkYSkpO1xuXG4gICAgLy8gSW50ZWdyYXRlIHRoZSBMYW1iZGEgZnVuY3Rpb25zIHdpdGggdGhlIEFQSSBHYXRld2F5IHJlc291cmNlXG4gICAgY29uc3QgZ2V0QWxsSW50ZWdyYXRpb24gPSBuZXcgTGFtYmRhSW50ZWdyYXRpb24oZ2V0QWxsTGFtYmRhKTtcblxuICAgIC8vIENyZWF0ZSBhbiBBUEkgR2F0ZXdheSByZXNvdXJjZSBmb3IgZWFjaCBvZiB0aGUgQ1JVRCBvcGVyYXRpb25zXG4gICAgY29uc3QgYXBpID0gbmV3IFJlc3RBcGkodGhpcywgJ2l0ZW1zQXBpJywge1xuICAgICAgcmVzdEFwaU5hbWU6ICdJdGVtcyBTZXJ2aWNlJ1xuICAgICAgLy8gSW4gY2FzZSB5b3Ugd2FudCB0byBtYW5hZ2UgYmluYXJ5IHR5cGVzLCB1bmNvbW1lbnQgdGhlIGZvbGxvd2luZ1xuICAgICAgLy8gYmluYXJ5TWVkaWFUeXBlczogW1wiKi8qXCJdLFxuICAgIH0pO1xuXG4gICAgY29uc3QgaXRlbXMgPSBhcGkucm9vdC5hZGRSZXNvdXJjZSgnaXRlbXMnKTtcbiAgICBpdGVtcy5hZGRNZXRob2QoJ0dFVCcsIGdldEFsbEludGVncmF0aW9uKTtcbiAgICBhZGRDb3JzT3B0aW9ucyhpdGVtcyk7XG4gIH1cbn1cblxuZXhwb3J0IGZ1bmN0aW9uIGFkZENvcnNPcHRpb25zKGFwaVJlc291cmNlOiBJUmVzb3VyY2UpIHtcbiAgYXBpUmVzb3VyY2UuYWRkTWV0aG9kKCdPUFRJT05TJywgbmV3IE1vY2tJbnRlZ3JhdGlvbih7XG4gICAgLy8gSW4gY2FzZSB5b3Ugd2FudCB0byB1c2UgYmluYXJ5IG1lZGlhIHR5cGVzLCB1bmNvbW1lbnQgdGhlIGZvbGxvd2luZyBsaW5lXG4gICAgLy8gY29udGVudEhhbmRsaW5nOiBDb250ZW50SGFuZGxpbmcuQ09OVkVSVF9UT19URVhULFxuICAgIGludGVncmF0aW9uUmVzcG9uc2VzOiBbe1xuICAgICAgc3RhdHVzQ29kZTogJzIwMCcsXG4gICAgICByZXNwb25zZVBhcmFtZXRlcnM6IHtcbiAgICAgICAgJ21ldGhvZC5yZXNwb25zZS5oZWFkZXIuQWNjZXNzLUNvbnRyb2wtQWxsb3ctSGVhZGVycyc6IFwiJ0NvbnRlbnQtVHlwZSxYLUFtei1EYXRlLEF1dGhvcml6YXRpb24sWC1BcGktS2V5LFgtQW16LVNlY3VyaXR5LVRva2VuLFgtQW16LVVzZXItQWdlbnQnXCIsXG4gICAgICAgICdtZXRob2QucmVzcG9uc2UuaGVhZGVyLkFjY2Vzcy1Db250cm9sLUFsbG93LU9yaWdpbic6IFwiJyonXCIsXG4gICAgICAgICdtZXRob2QucmVzcG9uc2UuaGVhZGVyLkFjY2Vzcy1Db250cm9sLUFsbG93LUNyZWRlbnRpYWxzJzogXCInZmFsc2UnXCIsXG4gICAgICAgICdtZXRob2QucmVzcG9uc2UuaGVhZGVyLkFjY2Vzcy1Db250cm9sLUFsbG93LU1ldGhvZHMnOiBcIidPUFRJT05TLEdFVCxQVVQsUE9TVCxERUxFVEUnXCIsXG4gICAgICB9LFxuICAgIH1dLFxuICAgIC8vIEluIGNhc2UgeW91IHdhbnQgdG8gdXNlIGJpbmFyeSBtZWRpYSB0eXBlcywgY29tbWVudCBvdXQgdGhlIGZvbGxvd2luZyBsaW5lXG4gICAgcGFzc3Rocm91Z2hCZWhhdmlvcjogUGFzc3Rocm91Z2hCZWhhdmlvci5ORVZFUixcbiAgICByZXF1ZXN0VGVtcGxhdGVzOiB7XG4gICAgICBcImFwcGxpY2F0aW9uL2pzb25cIjogXCJ7XFxcInN0YXR1c0NvZGVcXFwiOiAyMDB9XCJcbiAgICB9LFxuICB9KSwge1xuICAgIG1ldGhvZFJlc3BvbnNlczogW3tcbiAgICAgIHN0YXR1c0NvZGU6ICcyMDAnLFxuICAgICAgcmVzcG9uc2VQYXJhbWV0ZXJzOiB7XG4gICAgICAgICdtZXRob2QucmVzcG9uc2UuaGVhZGVyLkFjY2Vzcy1Db250cm9sLUFsbG93LUhlYWRlcnMnOiB0cnVlLFxuICAgICAgICAnbWV0aG9kLnJlc3BvbnNlLmhlYWRlci5BY2Nlc3MtQ29udHJvbC1BbGxvdy1NZXRob2RzJzogdHJ1ZSxcbiAgICAgICAgJ21ldGhvZC5yZXNwb25zZS5oZWFkZXIuQWNjZXNzLUNvbnRyb2wtQWxsb3ctQ3JlZGVudGlhbHMnOiB0cnVlLFxuICAgICAgICAnbWV0aG9kLnJlc3BvbnNlLmhlYWRlci5BY2Nlc3MtQ29udHJvbC1BbGxvdy1PcmlnaW4nOiB0cnVlLFxuICAgICAgfSxcbiAgICB9XVxuICB9KVxufVxuXG5jb25zdCBhcHAgPSBuZXcgQXBwKCk7XG5uZXcgQXBpTGFtYmRhQ3J1ZER5bmFtb0RCU3RhY2soYXBwLCAnc3ByaW5nZmllbGRNZWV0dXBBcGknKTtcbmFwcC5zeW50aCgpO1xuIl19