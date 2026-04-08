import * as bitcoin from "bitcoinjs-lib";
import crypto from "crypto";
import { ECPairFactory, ECPairInterface } from 'ecpair';
import * as tinysecp from 'tiny-secp256k1';
import { ElectrumClient } from "./electrum_client";
import btcPrivateKeys from "../../common/bitcoinkeys.json";
import { addressToElectrumScriptHash, selectUtxos } from "./utils";

const ECPair = ECPairFactory(tinysecp);

const electrumClient = new ElectrumClient("localhost", 50001);

async function main() {
    const destinationAddress = process.argv[2]
    const satoshisToSend = BigInt(process.argv[3] || "1000000");
    const txFee = 1000n

    const network = bitcoin.networks.regtest;

    const keyPair: ECPairInterface = ECPair.fromWIF(btcPrivateKeys.bip84[0].privatekey, network);

    const payment = bitcoin.payments.p2wpkh({ pubkey: keyPair.publicKey, network });
    const address = payment.address!

    const scriptPubKey = payment.output!

    const scriptHash = crypto.createHash('sha256')
        .update(scriptPubKey)
        .digest()
        .reverse()
        .toString('hex');


    const utxos = await electrumClient.listUnspentTxs(scriptHash)

    const { selected, totalValue } = selectUtxos(utxos, satoshisToSend, txFee);

    const psbt = new bitcoin.Psbt({ network });

    for (const utxo of selected) {        
        psbt.addInput({
            hash: utxo.tx_hash,
            index: utxo.tx_pos,
            witnessUtxo: {
                script: scriptPubKey,
                value: BigInt(utxo.value)
            }
        });
    }

    psbt.addOutput({
        address: destinationAddress,
        value: satoshisToSend,
    });

    const change = totalValue - satoshisToSend - txFee;
    if (change > 546) { // Dust threshold
        psbt.addOutput({
            address: address!,
            value: change,
        });
    }

    psbt.signAllInputs(keyPair);
    psbt.finalizeAllInputs();

    const rawTx = psbt.extractTransaction().toHex();
    const txid = await electrumClient.broadcastTx(rawTx)
    console.log(txid)
}

main().catch(console.error);