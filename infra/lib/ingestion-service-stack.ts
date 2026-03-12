import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2';
import * as logs from 'aws-cdk-lib/aws-logs';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as ssm from 'aws-cdk-lib/aws-ssm';
import { Construct } from 'constructs';

interface IngestionServiceStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  albSecurityGroup: ec2.SecurityGroup;
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
        POLL_ENABLED:     process.env.POLL_ENABLED     ?? 'false',
        POLL_INTERVAL:    process.env.POLL_INTERVAL    ?? '5m',
        NOAA_ENABLED:     process.env.NOAA_ENABLED     ?? 'false',
        NOAA_STATION_IDS: process.env.NOAA_STATION_IDS ?? '',
        OWM_ENABLED:      process.env.OWM_ENABLED      ?? 'false',
        OWM_LOCATIONS:    process.env.OWM_LOCATIONS    ?? '',
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

    // Internal ALB — only reachable via the VPC Link, not the public internet
    const alb = new elbv2.ApplicationLoadBalancer(this, 'IngestionAlb', {
      vpc: props.vpc,
      internetFacing: false,
      securityGroup: props.albSecurityGroup,
    });

    const targetGroup = new elbv2.ApplicationTargetGroup(this, 'IngestionTargetGroup', {
      vpc: props.vpc,
      port: 8080,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      targets: [this.service.loadBalancerTarget({ containerName: 'ingestion', containerPort: 8080 })],
      healthCheck: {
        path: '/healthz',
        healthyHttpCodes: '200',
      },
    });

    const albListener = alb.addListener('HttpListener', {
      port: 80,
      protocol: elbv2.ApplicationProtocol.HTTP,
      defaultTargetGroups: [targetGroup],
    });

    // Write listener ARN to SSM so EcosApi can reference it without a CFN export/import lock
    new ssm.StringParameter(this, 'AlbListenerArnParam', {
      parameterName: '/ecos/ingestion/alb-listener-arn',
      stringValue: albListener.listenerArn,
    });

    new cdk.CfnOutput(this, 'IngestionClusterName', {
      value: props.cluster.clusterName,
    });

    new cdk.CfnOutput(this, 'IngestionServiceName', {
      value: this.service.serviceName,
    });
  }
}
