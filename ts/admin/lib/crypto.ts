import path from "path";
import fs from "fs";
import {KEYUTIL} from 'jsrsasign';
import {PathString} from "@hyperledger-twgc/fabric-formatter/lib/fs/path";

export namespace jsrsasign {
    export interface PrivateKey {
        type: 'EC'
        curveName: 'secp256r1',
        isPrivate: true,
        isPublic: false,
        prvKeyHex: string,
        pubKeyHex: string,
    }
}

export class ECDSA_PrvKey {
    #key: jsrsasign.PrivateKey

    constructor(key: jsrsasign.PrivateKey) {
        this.#key = key;
    }

    pem() {
        return KEYUTIL.getPEM(this.#key, 'PKCS8PRV').trim();
    }

    toKeystore(dirName:PathString) {
        const filename = 'priv_sk';
        const absolutePath = path.resolve(dirName, filename);
        const data = this.pem();
        fs.mkdirSync(dirName, {recursive: true});
        fs.writeFileSync(absolutePath, data);

    }
}