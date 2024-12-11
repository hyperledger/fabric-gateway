import express, { Express } from 'express';
import * as http from 'node:http';

export interface ServerOptions {
    port: number;
    handlers: ((app: Express) => void)[];
}

export class Server {
    readonly #app = express();
    readonly #port: number;
    #server?: http.Server;

    constructor(options: ServerOptions) {
        this.#port = options.port;
        options.handlers.forEach((handler) => handler(this.#app));
    }

    start(): Promise<void> {
        return new Promise((resolve) => {
            this.#server = this.#app.listen(this.#port, resolve);
        });
    }

    stop(): Promise<void> {
        return new Promise((resolve, reject) => this.#server?.close((err) => (err ? resolve() : reject(err))));
    }
}
