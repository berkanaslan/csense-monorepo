package io.berkan;

import software.amazon.awscdk.Stack;
import software.amazon.awscdk.StackProps;
import software.amazon.awscdk.services.apigateway.LambdaIntegration;
import software.amazon.awscdk.services.apigateway.RestApi;
import software.amazon.awscdk.services.ec2.SubnetConfiguration;
import software.amazon.awscdk.services.ec2.SubnetSelection;
import software.amazon.awscdk.services.ec2.SubnetType;
import software.amazon.awscdk.services.ec2.Vpc;
import software.amazon.awscdk.services.iam.AnyPrincipal;
import software.amazon.awscdk.services.iam.Effect;
import software.amazon.awscdk.services.iam.PolicyStatement;
import software.amazon.awscdk.services.lambda.Architecture;
import software.amazon.awscdk.services.lambda.Code;
import software.amazon.awscdk.services.lambda.Function;
import software.amazon.awscdk.services.lambda.Runtime;
import software.amazon.awscdk.services.s3.BlockPublicAccess;
import software.amazon.awscdk.services.s3.Bucket;
import software.amazon.awscdk.services.s3.BucketAccessControl;
import software.constructs.Construct;

import java.util.Arrays;
import java.util.List;

// Defines ../apps/demo-parser stack
public class DemoParserStack extends Stack {
    protected static final String BUCKET_ID = "csense-demos";
    protected static final String VPC_ID = "csense-vpc";

    public DemoParserStack(final Construct scope, final String id, final StackProps props) {
        super(scope, id, props);

        // Create a new VPC with 2 subnets (public and private)
        final Vpc vpc = Vpc.Builder.create(this, VPC_ID)
                .subnetConfiguration(Arrays.asList(
                        SubnetConfiguration.builder()
                                .name("csense-public-subnet")
                                .subnetType(SubnetType.PUBLIC)
                                .cidrMask(24)
                                .build(),
                        SubnetConfiguration.builder()
                                .name("csense-private-subnet")
                                .subnetType(SubnetType.PRIVATE_ISOLATED)
                                .cidrMask(24)
                                .build()
                ))
                .maxAzs(1)
                .natGateways(1)
                .build();

        // Create a new Lambda function that triggers the DemoParser Lambda function
        final Function demoParser = Function.Builder.create(this, "DemoParser")
                .runtime(Runtime.GO_1_X) // TODO: Deprecated. Migrate to Runtime.AL2023
                .code(Code.fromAsset("../apps/demo-parser/output"))
                .handler("main")
                .architecture(Architecture.X86_64)
                .timeout(software.amazon.awscdk.Duration.minutes(3))
                .vpc(vpc)
                .vpcSubnets(SubnetSelection.builder()
                        .subnetType(SubnetType.PUBLIC)
                        .build())
                .allowPublicSubnet(true)
                .initialPolicy(List.of(
                                PolicyStatement.Builder.create()
                                        .effect(Effect.ALLOW)
                                        .actions(Arrays.asList("logs:CreateLogGroup",
                                                "logs:CreateLogStream",
                                                "logs:PutLogEvents",
                                                "s3:*"
                                        ))
                                        .resources(List.of("*"))
                                        .build()
                        )
                )
                .build();

        // Create a new REST API that triggers the DemoParser Lambda function
        final RestApi api = RestApi.Builder.create(this, "DemoParserRestApi")
                .restApiName("DemoParser Service")
                .description("This service triggers the DemoParser Lambda function.")
                .build();

        final LambdaIntegration lambdaIntegration = new LambdaIntegration(demoParser);
        api.getRoot().addMethod("POST", lambdaIntegration);

        // Create a new S3 bucket
        final Bucket csenseDemosBucket = Bucket.Builder.create(this, BUCKET_ID)
                .bucketName(BUCKET_ID)
                .blockPublicAccess(BlockPublicAccess.BLOCK_ACLS)
                .accessControl(BucketAccessControl.BUCKET_OWNER_FULL_CONTROL)
                .build();

        // Add a resource policy to the bucket
        csenseDemosBucket.addToResourcePolicy(PolicyStatement.Builder.create()
                .effect(Effect.ALLOW)
                .actions(List.of("s3:*"))
                .principals(List.of(new AnyPrincipal()))
                .resources(Arrays.asList(csenseDemosBucket.getBucketArn(), csenseDemosBucket.getBucketArn() + "/*"))
                .build());
    }
}
