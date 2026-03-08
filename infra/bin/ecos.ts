#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { NetworkStack } from '../lib/network-stack';
import { DatabaseStack } from '../lib/database-stack';
import { AuthStack } from '../lib/auth-stack';
import { MessagingStack } from '../lib/messaging-stack';
import { ComputeStack } from '../lib/compute-stack';
import { ApiStack } from '../lib/api-stack';
import { IngestionServiceStack } from '../lib/ingestion-service-stack';
import { FrontendStack } from '../lib/frontend-stack';
import { CacheStack } from '../lib/cache-stack';

const app = new cdk.App();

const env = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION ?? 'us-east-1',
};

// Foundation — no dependencies
const networkStack = new NetworkStack(app, 'EcosNetwork', { env });
const authStack = new AuthStack(app, 'EcosAuth', { env });
const messagingStack = new MessagingStack(app, 'EcosMessaging', { env });
const frontendStack = new FrontendStack(app, 'EcosFrontend', { env });

// Depends on: network
const databaseStack = new DatabaseStack(app, 'EcosDatabase', {
  env,
  vpc: networkStack.vpc,
  rdsSecurityGroup: networkStack.rdsSecurityGroup,
});

const computeStack = new ComputeStack(app, 'EcosCompute', {
  env,
  vpc: networkStack.vpc,
});

const cacheStack = new CacheStack(app, 'EcosCache', {
  env,
  vpc: networkStack.vpc,
  cacheSecurityGroup: networkStack.cacheSecurityGroup,
});

// Depends on: network, compute, database, messaging
const ingestionServiceStack = new IngestionServiceStack(app, 'EcosIngestionService', {
  env,
  vpc: networkStack.vpc,
  ecsSecurityGroup: networkStack.ecsSecurityGroup,
  cluster: computeStack.cluster,
  ingestionLogGroup: computeStack.ingestionLogGroup,
  sensorReadingsTable: databaseStack.sensorReadingsTable,
  stationsTable: databaseStack.stationsTable,
  dataIngestionTopic: messagingStack.dataIngestionTopic,
  alertEventsTopic: messagingStack.alertEventsTopic,
});

// Depends on: auth, network, ingestion service
const apiStack = new ApiStack(app, 'EcosApi', {
  env,
  userPool: authStack.userPool,
  userPoolClient: authStack.userPoolClient,
  vpc: networkStack.vpc,
  ecsSecurityGroup: networkStack.ecsSecurityGroup,
  ingestionService: ingestionServiceStack.service,
});

app.synth();
