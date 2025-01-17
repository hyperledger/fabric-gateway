import JsdomEnvironment from 'jest-environment-jsdom';
import { EnvironmentContext, JestEnvironmentConfig } from '@jest/environment';
import { webcrypto } from 'crypto';

class TestEnvironment extends JsdomEnvironment {
    constructor(config: JestEnvironmentConfig, context: EnvironmentContext) {
        super(config, context);

        this.global.Uint8Array = Uint8Array;
        this.global.TextEncoder = TextEncoder;
        this.global.TextDecoder = TextDecoder;
        Object.defineProperty(this.global, 'crypto', {
            value: webcrypto,
        });
    }
}

export default TestEnvironment;
