import * as cdk from 'aws-cdk-lib';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import { Construct } from 'constructs';

interface ApiStackProps extends cdk.StackProps {
  userPool: cognito.UserPool;
  userPoolClient: cognito.UserPoolClient;
  cluster: ecs.Cluster;
}

export class ApiStack extends cdk.Stack {
  public readonly api: apigatewayv2.CfnApi;
  public readonly apiEndpoint: string;

  constructor(scope: Construct, id: string, props: ApiStackProps) {
    super(scope, id, props);

    // HTTP API Gateway (cheaper than REST API)
    this.api = new apigatewayv2.CfnApi(this, 'EcosApi', {
      name: 'ecos-api',
      protocolType: 'HTTP',
      corsConfiguration: {
        allowOrigins: ['*'],
        allowMethods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        allowHeaders: ['Content-Type', 'Authorization'],
      },
    });

    // Cognito JWT authorizer
    const authorizer = new apigatewayv2.CfnAuthorizer(this, 'CognitoAuthorizer', {
      apiId: this.api.ref,
      authorizerType: 'JWT',
      name: 'cognito-authorizer',
      identitySource: ['$request.header.Authorization'],
      jwtConfiguration: {
        audience: [props.userPoolClient.userPoolClientId],
        issuer: `https://cognito-idp.${this.region}.amazonaws.com/${props.userPool.userPoolId}`,
      },
    });

    // Default stage with auto-deploy
    const stage = new apigatewayv2.CfnStage(this, 'DefaultStage', {
      apiId: this.api.ref,
      stageName: '$default',
      autoDeploy: true,
    });

    // Route stubs — these will be connected to real integrations
    // when the Go/Python services are deployed to ECS

    // Mock integration for placeholder routes
    const mockIntegration = new apigatewayv2.CfnIntegration(this, 'MockIntegration', {
      apiId: this.api.ref,
      integrationType: 'HTTP_PROXY',
      integrationMethod: 'GET',
      // Placeholder URI — will be replaced with ECS service ALB
      integrationUri: 'https://httpbin.org/anything',
    });

    const routes = [
      { path: '/api/readings', methods: ['GET', 'POST'] },
      { path: '/api/stations', methods: ['GET', 'POST'] },
      { path: '/api/analysis', methods: ['GET', 'POST'] },
    ];

    for (const route of routes) {
      for (const method of route.methods) {
        new apigatewayv2.CfnRoute(this, `Route-${method}-${route.path.replace(/\//g, '-')}`, {
          apiId: this.api.ref,
          routeKey: `${method} ${route.path}`,
          authorizationType: 'JWT',
          authorizerId: authorizer.ref,
          target: `integrations/${mockIntegration.ref}`,
        });
      }
    }

    this.apiEndpoint = `https://${this.api.ref}.execute-api.${this.region}.amazonaws.com`;

    new cdk.CfnOutput(this, 'ApiEndpoint', {
      value: this.apiEndpoint,
      description: 'ECOS API Gateway endpoint URL',
    });
  }
}
