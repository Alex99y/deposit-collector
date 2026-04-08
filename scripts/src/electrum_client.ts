import * as net from 'node:net';
import { addressToElectrumScriptHash } from './utils';

interface ElectrumResponse {
    jsonrpc: string;
    result?: any;
    error?: { code: number; message: string };
    id: number | string;
}

export interface UTXO {
    tx_hash: string;
    tx_pos: number;
    value: number;
    height: number;
}

export interface Balance {
    confirmed: number;
    unconfirmed: number;
}

export class ElectrumClient {
    private host: string;
    private port: number;

    constructor(host: string, port: number) {
        this.host = host;
        this.port = port;
    }

    public async getAddressBalance(address: string): Promise<Balance> {
        const publicKeyHash = addressToElectrumScriptHash(address);
        const balance = await this.request(
            "blockchain.scripthash.get_balance", [publicKeyHash]
        );

        return balance
    }

    public async listUnspentTxs(scriptHash: string): Promise<UTXO[]> {
        const utxos: UTXO[] = await this.request('blockchain.scripthash.listunspent', [scriptHash]);
        return utxos;
    }

    public async getTx (txHash: string): Promise<string> {
        const txHex: string = await this.request('blockchain.transaction.get', [txHash]);
        return txHex;
    }

    public async broadcastTx (rawTxHex: string): Promise<any> {
        const txid = await this.request('blockchain.transaction.broadcast', [rawTxHex]);
        return txid;
    }

    private async request(method: string, params: any[] = []): Promise<any> {
        return new Promise((resolve, reject) => {
            const client = new net.Socket();

            client.setTimeout(5000);

            client.connect(this.port, this.host, () => {
                const payload = JSON.stringify({
                    id: Date.now(),
                    method: method,
                    params: params
                }) + '\n';
                
                client.write(payload);
            });

            let responseData = ''

            const resolveData = function (result: string) {
                const response: ElectrumResponse = JSON.parse(result);
                    
                if (response.error) {
                    reject(response.error);
                } else {
                    resolve(response.result);
                }
                
                client.destroy();
            }

            client.on('data', (data) => {
                try {
                    if (typeof data === "string") {
                        resolveData(data.toString())
                    } else {
                        responseData += data.toString()
                        if (responseData.endsWith('\n')) {
                            resolveData(responseData)
                        }
                    }
                } catch (err) {
                    reject(new Error("Error parsing response"));
                }
            });

            client.on('error', (err) => {
                reject(err);
            });

            client.on('timeout', () => {
                client.destroy();
                reject(new Error("Timeout conectando a ElectrumX"));
            });
        });
    }
}