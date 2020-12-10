module.exports = {
    "roots": [
        "<rootDir>/src"
    ],
    "preset": "ts-jest",
    "testEnvironment": "node",
    "collectCoverage": true,
    "collectCoverageFrom": [
        "**/*.[jt]s?(x)",
        "!**/*.d.ts",
        "!src/protos/**",
    ],
    "coverageProvider": "v8",
    "testMatch": [
        "**/?(*.)+(spec|test).[jt]s?(x)"
    ],
}
