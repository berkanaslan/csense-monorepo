package io.berkan;

import software.amazon.awscdk.Stack;
import software.amazon.awscdk.StackProps;
import software.amazon.awscdk.services.apigateway.LambdaIntegration;
import software.amazon.awscdk.services.apigateway.RestApi;
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

    public DemoParserStack(final Construct scope, final String id, final StackProps props) {
        super(scope, id, props);

        // Create a new Lambda function that triggers the DemoParser Lambda function
        final Function demoParser = Function.Builder.create(this, "DemoParser")
                .runtime(Runtime.PROVIDED_AL2)
                .code(Code.fromAsset("../apps/demo-parser/output"))
                .handler("bootstrap")
                .architecture(Architecture.X86_64)
                .timeout(software.amazon.awscdk.Duration.minutes(15))
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
                .versioned(true)
                .blockPublicAccess(BlockPublicAccess.BLOCK_ALL)
                .accessControl(BucketAccessControl.PRIVATE)
                .build();

        csenseDemosBucket.grantReadWrite(demoParser);
    }
}
