import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as sns from 'aws-cdk-lib/aws-sns';
import { Construct } from 'constructs';

interface IngestionServiceStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  ecsSecurityGroup: ec2.SecurityGroup;
  cluster: ecs.Cluster;
  ingestionLogGroup: logs.LogGroup;
  sensorReadingsTable: dynamodb.Table;
  stationsTable: dynamodb.Table;
  dataIngestionTopic: sns.Topic;
  alertEventsTopic: sns.Topic;
  repository: ecr.Repository;
}

export class IngestionServiceStack extends cdk.Stack {
  public readonly service: ecs.FargateService;

  constructor(scope: Construct, id: string, props: IngestionServiceStackProps) {
    super(scope, id, props);

    const taskDefinition = new ecs.FargateTaskDefinition(this, 'IngestionTaskDef', {
      cpu: 256,
      memoryLimitMiB: 512,
    });

    props.repository.grantPull(taskDefinition.taskRole);

    taskDefinition.addContainer('ingestion', {
      image: ecs.ContainerImage.fromEcrRepository(props.repository, this.node.tryGetContext('imageTag') ?? 'latest'),
      portMappings: [{ containerPort: 8080 }],
      environment: {
        READINGS_TABLE_NAME: props.sensorReadingsTable.tableName,
        STATIONS_TABLE_NAME: props.stationsTable.tableName,
        READING_TYPE_INDEX_NAME: 'byReadingType',
        DATA_INGESTION_TOPIC_ARN: props.dataIngestionTopic.topicArn,
        ALERT_EVENTS_TOPIC_ARN: props.alertEventsTopic.topicArn,
        AWS_REGION: this.region,
        PORT: '8080',
      },
      logging: ecs.LogDriver.awsLogs({
        logGroup: props.ingestionLogGroup,
        streamPrefix: 'ecs',
      }),
      healthCheck: {
        command: ['CMD-SHELL', 'wget -qO- http://localhost:8080/healthz || exit 1'],
        interval: cdk.Duration.seconds(30),
        timeout: cdk.Duration.seconds(5),
        retries: 3,
        startPeriod: cdk.Duration.seconds(10),
      },
    });

    // IAM grants
    props.sensorReadingsTable.grantReadWriteData(taskDefinition.taskRole);
    props.stationsTable.grantReadWriteData(taskDefinition.taskRole);
    props.dataIngestionTopic.grantPublish(taskDefinition.taskRole);
    props.alertEventsTopic.grantPublish(taskDefinition.taskRole);

    this.service = new ecs.FargateService(this, 'IngestionService', {
      cluster: props.cluster,
      taskDefinition,
      desiredCount: 1,
      assignPublicIp: true,
      securityGroups: [props.ecsSecurityGroup],
      vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
      cloudMapOptions: {
        name: 'ingestion',
      },
      circuitBreaker: { rollback: false },
    });

    new cdk.CfnOutput(this, 'IngestionClusterName', {
      value: props.cluster.clusterName,
    });

    new cdk.CfnOutput(this, 'IngestionServiceName', {
      value: this.service.serviceName,
    });
  }
}
