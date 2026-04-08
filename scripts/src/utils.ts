import * as bitcoin from 'bitcoinjs-lib';
import * as crypto from 'crypto';
import { UTXO } from './electrum_client';

const REGTEST_NETWORK = bitcoin.networks.regtest

export function addressToElectrumScriptHash(address: string): string {
    const script = bitcoin.address.toOutputScript(address, REGTEST_NETWORK);
    const hash = crypto.createHash('sha256').update(script).digest();

    return hash.reverse().toString('hex');
}

export function selectUtxos(utxos: UTXO[], targetAmount: bigint, fee: bigint): { selected: UTXO[], totalValue: bigint } {
    const sortedUtxos = [...utxos].sort((a, b) => a.height - b.height);
    
    let totalValue = 0n;
    const selected: UTXO[] = [];

    for (const utxo of sortedUtxos) {
        selected.push(utxo);
        totalValue += BigInt(utxo.value);
        if (totalValue >= targetAmount + fee) break;
    }

    if (totalValue < targetAmount + fee) {
        throw new Error("Not enough funds to send the transaction");
    }

    return { selected, totalValue };
}