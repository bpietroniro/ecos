import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

interface DatabaseStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  rdsSecurityGroup: ec2.SecurityGroup;
}

export class DatabaseStack extends cdk.Stack {
  public readonly sensorReadingsTable: dynamodb.Table;
  public readonly stationsTable: dynamodb.Table;

  constructor(scope: Construct, id: string, props: DatabaseStackProps) {
    super(scope, id, props);

    // DynamoDB — SensorReadings table
    this.sensorReadingsTable = new dynamodb.Table(this, 'SensorReadings', {
      tableName: 'ecos-sensor-readings',
      partitionKey: { name: 'stationId', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'sk', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    this.sensorReadingsTable.addGlobalSecondaryIndex({
      indexName: 'byReadingType',
      partitionKey: { name: 'readingType', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'sk', type: dynamodb.AttributeType.STRING },
      projectionType: dynamodb.ProjectionType.ALL,
    });

    // DynamoDB — Stations table
    this.stationsTable = new dynamodb.Table(this, 'Stations', {
      tableName: 'ecos-stations',
      partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // TODO: Phase 2 — RDS PostgreSQL for analysis results
    // Will use props.vpc and props.rdsSecurityGroup
    // const analysisDb = new rds.DatabaseInstance(this, 'AnalysisDb', {
    //   engine: rds.DatabaseInstanceEngine.postgres({ version: rds.PostgresEngineVersion.VER_16 }),
    //   instanceType: ec2.InstanceType.of(ec2.InstanceClass.T4G, ec2.InstanceSize.MICRO),
    //   vpc: props.vpc,
    //   vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
    //   securityGroups: [props.rdsSecurityGroup],
    //   databaseName: 'ecos_analysis',
    //   allocatedStorage: 20,
    //   removalPolicy: cdk.RemovalPolicy.DESTROY,
    // });
  }
}
