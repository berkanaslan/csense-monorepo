package io.berkan;

import software.amazon.awscdk.App;
import software.amazon.awscdk.Environment;
import software.amazon.awscdk.StackProps;

public class CSenseApp {
    static Environment getEnvironment() {
        final String account = System.getenv("CDK_DEFAULT_ACCOUNT");
        final String region = System.getenv("CDK_DEFAULT_REGION");

        return Environment.builder()
                .account(account)
                .region(region)
                .build();
    }

    public static void main(final String[] args) {
        final App app = new App();
        final Environment env = getEnvironment();

        // Demo Parser
        final StackProps csenseStackProperties = StackProps.builder()
                .env(env)
                .build();

        final DemoParserStack csenseStack = new DemoParserStack(app, "CSenseStack", csenseStackProperties);

        // TODO: More stacks can be added here...
        // ...

        app.synth();
    }
}

