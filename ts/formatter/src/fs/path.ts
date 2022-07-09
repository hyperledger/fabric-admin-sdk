import fs from 'fs';
import path from 'path';
import {normalizeX509} from '../helper'
export function findKeyFiles (dir)  {
    const files = fs.readdirSync(dir);
    return files.filter((fileName) => fileName.endsWith('_sk')).map((fileName) => path.resolve(dir, fileName));
}

export function findCertFiles (dir) {
    const files = fs.readdirSync(dir);
    return files.map((fileName) => path.resolve(dir, fileName)).filter(filePath => {
        try {
            const pem = fs.readFileSync(filePath).toString();
            normalizeX509(pem);
            return true;
        } catch (e) {
            return false;
        }
    });
}


export class ECDSA_PrvKey {
    #key
    constructor(key) {
        // TODO how is this type?
        this.#key = key;
    }

    pem() {
        return this.#key.toBytes();
    }

    toKeystore(dirName) {
        const filename = 'priv_sk';
        const absolutePath = path.resolve(dirName, filename);
        const data = this.pem();
        fs.mkdirSync(dirName, {recursive: true});
        fs.writeFileSync(absolutePath, data);

    }
}
