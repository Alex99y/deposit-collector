import { ElectrumClient } from "./electrum_client";
import btcPrivateKeys from "../../common/bitcoinkeys.json";

const electrumClient = new ElectrumClient("localhost", 50001);

async function main() {
    const address = process.argv[2] || btcPrivateKeys.bip84[0].address;
    const balance = await electrumClient.getAddressBalance(address)
    console.log(balance);
}

main().catch(console.error);