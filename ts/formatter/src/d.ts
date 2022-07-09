export type ValueOf<T> = T[keyof T];
export type IndexDigit = number

export function isIndexDigit(num): num is IndexDigit {
    return Number.isInteger(num) && num >= 0
}

/**
 * MSPName maps to MSP ID, and usually to organization name
 */
export type MSPName = string

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