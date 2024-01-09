import fs from 'fs';
import path from 'path';
import {normalizeX509} from '../helper'

export type PathString = string

export function findKeyFiles(dir: PathString) {
    const files = fs.readdirSync(dir);
    return files.filter((fileName) => fileName.endsWith('_sk')).map((fileName) => path.resolve(dir, fileName));
}

export function findCertFiles(dir: PathString) {
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
