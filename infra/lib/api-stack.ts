import * as cdk from 'aws-cdk-lib';
import * as apigatewayv2 from 'aws-cdk-lib/aws-apigatewayv2';
import * as cognito from 'aws-cdk-lib/aws-cognito';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ssm from 'aws-cdk-lib/aws-ssm';
import { Construct } from 'constructs';

interface ApiStackProps extends cdk.StackProps {
  userPool: cognito.UserPool;
  userPoolClient: cognito.UserPoolClient;
  vpc: ec2.Vpc;
  vpcLinkSecurityGroup: ec2.SecurityGroup;
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
        // TODO: restrict to the frontend domain once it's known (e.g. CloudFront URL)
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
    new apigatewayv2.CfnStage(this, 'DefaultStage', {
      apiId: this.api.ref,
      stageName: '$default',
      autoDeploy: true,
    });

    // VPC Link for private integration — ENIs use the dedicated VPC Link SG
    const vpcLink = new apigatewayv2.CfnVpcLink(this, 'VpcLink', {
      name: 'ecos-vpc-link',
      subnetIds: props.vpc.publicSubnets.map(s => s.subnetId),
      securityGroupIds: [props.vpcLinkSecurityGroup.securityGroupId],
    });

    // Integration to ingestion service via internal ALB listener
    const ingestionIntegration = new apigatewayv2.CfnIntegration(this, 'IngestionIntegration', {
      apiId: this.api.ref,
      integrationType: 'HTTP_PROXY',
      integrationMethod: 'ANY',
      connectionType: 'VPC_LINK',
      connectionId: vpcLink.ref,
      integrationUri: ssm.StringParameter.valueForStringParameter(this, '/ecos/ingestion/alb-listener-arn'),
      payloadFormatVersion: '1.0',
    });

    // Routes — health check (no auth)
    new apigatewayv2.CfnRoute(this, 'Route-GET-healthz', {
      apiId: this.api.ref,
      routeKey: 'GET /healthz',
      target: `integrations/${ingestionIntegration.ref}`,
    });

    // Routes — ingestion service (JWT auth)
    const authedRoutes = [
      'GET /api/readings',
      'GET /api/readings/station/{stationId}',
      'POST /api/readings',
      'POST /api/readings/batch',
      'GET /api/stations',
      'POST /api/stations',
      'PUT /api/stations/{id}',
      'DELETE /api/stations/{id}',
    ];

    for (const routeKey of authedRoutes) {
      const routeId = routeKey.replace(/[\s\/\{\}]/g, '-');
      new apigatewayv2.CfnRoute(this, `Route-${routeId}`, {
        apiId: this.api.ref,
        routeKey,
        authorizationType: 'JWT',
        authorizerId: authorizer.ref,
        target: `integrations/${ingestionIntegration.ref}`,
      });
    }

    this.apiEndpoint = `https://${this.api.ref}.execute-api.${this.region}.amazonaws.com`;

    new cdk.CfnOutput(this, 'ApiEndpoint', {
      value: this.apiEndpoint,
      description: 'ECOS API Gateway endpoint URL',
    });
  }
}
