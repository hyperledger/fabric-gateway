import { Gateway } from '@hyperledger/fabric-gateway';
import express, { Express } from 'express';
import { Server } from './server';

const REST_PORT = 3000;

export interface GatewayServerOptions {
    port: number;
    gateway: Gateway;
}

export class GatewayServer {
    #gateway: Gateway;
    #server: Server;

    constructor(options: GatewayServerOptions) {
        this.#gateway = options.gateway;
        this.#server = new Server({
            port: options.port,
            handlers: [this.#evaluate],
        });
    }

    start(): Promise<void> {
        return this.#server.start();
    }

    stop(): Promise<void> {
        return this.#server.stop();
    }

    #evaluate(app: Express): void {
        app.post('/evaluate', express.json(), (request, response) => {
            request.body.proposal;
        });
    }
}
