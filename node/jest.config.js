module.exports = {
    "roots": [
        "<rootDir>/src"
    ],
    'preset': 'ts-jest',
    'testEnvironment': 'node',
    'collectCoverage': true,
    'collectCoverageFrom': [
        '**/*.[jt]s?(x)',
        '!**/*.d.ts',
    ],
    'coverageProvider': 'v8',
    'testMatch': [
        '**/?(*.)+(spec|test).[jt]s?(x)'
    ],
    'maxWorkers': 1, // Workaround for Jest BigInt serialization bug: https://github.com/facebook/jest/issues/11617
}
