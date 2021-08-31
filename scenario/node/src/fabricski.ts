import * as crypto from 'crypto';
import * as jsrsa from 'jsrsasign';

export function getSKIFromCertificate(pem: string): Buffer {
    const key = jsrsa.KEYUTIL.getKey(pem);
    const uncompressedPoint = getUncompressedPointOnCurve(key as jsrsa.KJUR.crypto.ECDSA);
    const hashBuffer = crypto.createHash('sha256');
    hashBuffer.update(uncompressedPoint);

    const digest = hashBuffer.digest('hex');
    return Buffer.from(digest, 'hex');
}

function getUncompressedPointOnCurve(key: jsrsa.KJUR.crypto.ECDSA): Buffer {
    const xyhex = key.getPublicKeyXYHex();
    const xBuffer = Buffer.from(xyhex.x, 'hex');
    const yBuffer = Buffer.from(xyhex.y, 'hex');
    const uncompressedPrefix = Buffer.from('04', 'hex');
    const uncompressedPoint = Buffer.concat([uncompressedPrefix, xBuffer, yBuffer]);
    return uncompressedPoint;
}
