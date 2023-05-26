/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import org.bouncycastle.asn1.ASN1Integer;

import java.math.BigInteger;
import java.security.GeneralSecurityException;
import java.security.interfaces.ECPrivateKey;

final class ECPrivateKeySigner implements Signer {
    private static final String ALGORITHM_NAME = "NONEwithECDSA";

    private final Signer signer;
    private final BigInteger curveN;
    private final BigInteger halfCurveN;

    ECPrivateKeySigner(final ECPrivateKey privateKey) {
        signer = new PrivateKeySigner(privateKey, ALGORITHM_NAME);
        curveN = privateKey.getParams().getOrder();
        halfCurveN = curveN.divide(BigInteger.valueOf(2));
    }

    @Override
    public byte[] sign(final byte[] digest) throws GeneralSecurityException {
        byte[] rawSignature = signer.sign(digest);
        ECSignature signature = ECSignature.fromBytes(rawSignature);
        signature = preventMalleability(signature);
        return signature.getBytes();
    }

    private ECSignature preventMalleability(final ECSignature signature) {
        BigInteger s = signature.getS().getValue();
        if (s.compareTo(halfCurveN) > 0) {
            s = curveN.subtract(s);
            return new ECSignature(signature.getR(), new ASN1Integer(s));
        }
        return signature;
    }
}
