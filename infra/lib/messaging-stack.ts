import * as cdk from 'aws-cdk-lib';
import * as sns from 'aws-cdk-lib/aws-sns';
import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as subscriptions from 'aws-cdk-lib/aws-sns-subscriptions';
import { Construct } from 'constructs';

export class MessagingStack extends cdk.Stack {
  public readonly dataIngestionTopic: sns.Topic;
  public readonly analysisEventsTopic: sns.Topic;
  public readonly alertEventsTopic: sns.Topic;
  public readonly analysisProcessingQueue: sqs.Queue;
  public readonly alertQueue: sqs.Queue;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // SNS Topics
    this.dataIngestionTopic = new sns.Topic(this, 'DataIngestionEvents', {
      topicName: 'ecos-data-ingestion-events',
    });

    this.analysisEventsTopic = new sns.Topic(this, 'AnalysisEvents', {
      topicName: 'ecos-analysis-events',
    });

    this.alertEventsTopic = new sns.Topic(this, 'AlertEvents', {
      topicName: 'ecos-alert-events',
    });

    // SQS Queues with Dead Letter Queues

    const analysisProcessingDlq = new sqs.Queue(this, 'AnalysisProcessingDLQ', {
      queueName: 'ecos-analysis-processing-dlq',
      retentionPeriod: cdk.Duration.days(14),
    });

    this.analysisProcessingQueue = new sqs.Queue(this, 'AnalysisProcessingQueue', {
      queueName: 'ecos-analysis-processing',
      visibilityTimeout: cdk.Duration.seconds(300),
      deadLetterQueue: {
        queue: analysisProcessingDlq,
        maxReceiveCount: 3,
      },
    });

    const alertDlq = new sqs.Queue(this, 'AlertDLQ', {
      queueName: 'ecos-alert-dlq',
      retentionPeriod: cdk.Duration.days(14),
    });

    this.alertQueue = new sqs.Queue(this, 'AlertQueue', {
      queueName: 'ecos-alert-queue',
      visibilityTimeout: cdk.Duration.seconds(60),
      deadLetterQueue: {
        queue: alertDlq,
        maxReceiveCount: 3,
      },
    });

    // Subscriptions: SNS → SQS
    this.dataIngestionTopic.addSubscription(
      new subscriptions.SqsSubscription(this.analysisProcessingQueue),
    );

    this.alertEventsTopic.addSubscription(
      new subscriptions.SqsSubscription(this.alertQueue),
    );
  }
}
