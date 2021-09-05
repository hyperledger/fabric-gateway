/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.math.BigInteger;
import java.security.GeneralSecurityException;
import java.security.Provider;
import java.security.Signature;
import java.security.interfaces.ECPrivateKey;

import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.jce.provider.BouncyCastleProvider;

final class ECPrivateKeySigner implements Signer {
    private static final Provider PROVIDER = new BouncyCastleProvider();
    private static final String ALGORITHM_NAME = "NONEwithECDSA";

    private final ECPrivateKey privateKey;
    private final BigInteger curveN;
    private final BigInteger halfCurveN;

    ECPrivateKeySigner(final ECPrivateKey privateKey) {
        this.privateKey = privateKey;
        curveN = privateKey.getParams().getOrder();
        halfCurveN = curveN.divide(BigInteger.valueOf(2));
    }

    @Override
    public byte[] sign(final byte[] digest) throws GeneralSecurityException {
        byte[] rawSignature = generateSignature(digest);
        ECSignature signature = ECSignature.fromBytes(rawSignature);
        signature = preventMalleability(signature);
        return signature.getBytes();
    }

    private byte[] generateSignature(final byte[] digest) throws GeneralSecurityException {
        Signature signer = Signature.getInstance(ALGORITHM_NAME, PROVIDER);
        signer.initSign(privateKey);
        signer.update(digest);
        return signer.sign();
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
