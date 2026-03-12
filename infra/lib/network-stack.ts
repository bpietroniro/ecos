import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

export class NetworkStack extends cdk.Stack {
  public readonly vpc: ec2.Vpc;
  public readonly vpcLinkSecurityGroup: ec2.SecurityGroup;
  public readonly albSecurityGroup: ec2.SecurityGroup;
  public readonly ecsSecurityGroup: ec2.SecurityGroup;
  public readonly rdsSecurityGroup: ec2.SecurityGroup;
  public readonly cacheSecurityGroup: ec2.SecurityGroup;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // VPC with public subnets only — no NAT gateway to save ~$30/month
    this.vpc = new ec2.Vpc(this, 'EcosVpc', {
      maxAzs: 2,
      natGateways: 0,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
        },
      ],
    });

    // Security group for the API Gateway VPC Link ENIs
    this.vpcLinkSecurityGroup = new ec2.SecurityGroup(this, 'VpcLinkSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for API Gateway VPC Link',
      allowAllOutbound: false,
    });

    // Security group for the internal ALB — only accepts traffic from VPC Link
    this.albSecurityGroup = new ec2.SecurityGroup(this, 'AlbSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for internal ALB',
      allowAllOutbound: false,
    });
    this.albSecurityGroup.addIngressRule(
      this.vpcLinkSecurityGroup,
      ec2.Port.tcp(80),
      'Allow inbound HTTP from VPC Link',
    );

    // Security group for ECS Fargate services — only accepts traffic from ALB
    this.ecsSecurityGroup = new ec2.SecurityGroup(this, 'EcsSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for ECS Fargate services',
      allowAllOutbound: true,
    });
    this.ecsSecurityGroup.addIngressRule(
      this.albSecurityGroup,
      ec2.Port.tcp(8080),
      'Allow inbound from ALB',
    );

    // Explicit egress: VPC Link → ALB on port 80 only
    this.vpcLinkSecurityGroup.addEgressRule(
      this.albSecurityGroup,
      ec2.Port.tcp(80),
      'Allow outbound to ALB',
    );

    // Explicit egress: ALB → ECS on port 8080 only (listener + health checks)
    this.albSecurityGroup.addEgressRule(
      this.ecsSecurityGroup,
      ec2.Port.tcp(8080),
      'Allow outbound to ECS targets',
    );

    // Security group for RDS PostgreSQL (Phase 2)
    this.rdsSecurityGroup = new ec2.SecurityGroup(this, 'RdsSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for RDS PostgreSQL',
      allowAllOutbound: false,
    });
    this.rdsSecurityGroup.addIngressRule(
      this.ecsSecurityGroup,
      ec2.Port.tcp(5432),
      'Allow PostgreSQL access from ECS services',
    );

    // Security group for ElastiCache Redis (Phase 2)
    this.cacheSecurityGroup = new ec2.SecurityGroup(this, 'CacheSecurityGroup', {
      vpc: this.vpc,
      description: 'Security group for ElastiCache Redis',
      allowAllOutbound: false,
    });
    this.cacheSecurityGroup.addIngressRule(
      this.ecsSecurityGroup,
      ec2.Port.tcp(6379),
      'Allow Redis access from ECS services',
    );
  }
}
