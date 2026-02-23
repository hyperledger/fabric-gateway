export default {
    roots: ['<rootDir>/src'],
    extensionsToTreatAsEsm: ['.ts'],
    moduleNameMapper: { '^(\\.{1,2}/.*)\\.js$': '$1' },
    transform: {
        '^.+\\.tsx?$': ['ts-jest', { useESM: true }],
        '^.+\\.js$': ['ts-jest', { useESM: true }],
    },
    transformIgnorePatterns: ['node_modules/(?!(@noble)/)'],
    testEnvironment: 'node',
    collectCoverage: true,
    collectCoverageFrom: ['**/*.[jt]s?(x)', '!**/*.d.ts'],
    coverageProvider: 'v8',
    testMatch: ['**/?(*.)+(spec|test).[jt]s?(x)'],
    verbose: true,
    workerThreads: true,
};
