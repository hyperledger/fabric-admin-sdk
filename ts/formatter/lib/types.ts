export type ValueOf<T> = T[keyof T];
export type IndexDigit = number

export function isIndex(num: number): num is IndexDigit {
    return Number.isInteger(num) && num >= 0
}

/**
 * Transaction Id
 */
export type TxId = string

export interface ConnectionInfo {
    type: string;
    url: string;
    options: Record<string, unknown>;
}

export interface WithConnectionInfo {
    connection: ConnectionInfo;
}

/**
 * MSPName maps to MSP ID, and usually to organization name
 */
export type MSPName = string


export type ChaincodeLabel = string
export type ChannelName = string

/**
 * MspId is short form of _Member Service Provider Identifier_
 */
export type MspId = string

/**
 * PEM format string
 */
export type PEM = string

/**
 * certificate containing the public key in PEM format.
 */
export type CertificatePEM = PEM