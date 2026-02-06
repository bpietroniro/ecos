import * as cdk from 'aws-cdk-lib';
import * as ec2 from 'aws-cdk-lib/aws-ec2';
import { Construct } from 'constructs';

interface CacheStackProps extends cdk.StackProps {
  vpc: ec2.Vpc;
  cacheSecurityGroup: ec2.SecurityGroup;
}

export class CacheStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CacheStackProps) {
    super(scope, id, props);

    // TODO: Phase 2 — ElastiCache Redis for caching analysis results
    // Deferred to avoid idle costs (~$13/month for cache.t4g.micro)
    //
    // const subnetGroup = new elasticache.CfnSubnetGroup(this, 'RedisSubnetGroup', {
    //   description: 'Subnet group for ECOS Redis cluster',
    //   subnetIds: props.vpc.publicSubnets.map(s => s.subnetId),
    //   cacheSubnetGroupName: 'ecos-redis-subnet-group',
    // });
    //
    // const redisCluster = new elasticache.CfnCacheCluster(this, 'RedisCluster', {
    //   engine: 'redis',
    //   cacheNodeType: 'cache.t4g.micro',
    //   numCacheNodes: 1,
    //   clusterName: 'ecos-redis',
    //   vpcSecurityGroupIds: [props.cacheSecurityGroup.securityGroupId],
    //   cacheSubnetGroupName: subnetGroup.ref,
    // });
  }
}
