/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.io.ByteArrayInputStream;
import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.UncheckedIOException;
import java.math.BigInteger;
import java.security.GeneralSecurityException;
import java.security.Provider;
import java.security.Signature;
import java.security.interfaces.ECPrivateKey;
import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.StreamSupport;

import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DERSequenceGenerator;
import org.bouncycastle.asn1.x9.ECNamedCurveTable;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.jce.provider.BouncyCastleProvider;

/**
 * A signing implementation used to generate digital signatures.
 */
final class ECDSAPrivateKeySigner implements Signer {
    private static final Provider PROVIDER = new BouncyCastleProvider();
    private static final X9ECParameters CURVE = ECNamedCurveTable.getByName("P-256");
    private static final BigInteger HALF_CURVE_N = CURVE.getN().divide(BigInteger.valueOf(2));
    private static final String ALGORITHM_NAME = "NONEwithECDSA";

    private final ECPrivateKey privateKey;

    private static final class ECSignature {
        public final ASN1Integer r;
        public final ASN1Integer s;

        public ECSignature(final ASN1Integer r, final ASN1Integer s) {
            this.r = r;
            this.s = s;
        }

        public byte[] getBytes() {
            try (ByteArrayOutputStream bytesOut = new ByteArrayOutputStream()) {
                DERSequenceGenerator seq = new DERSequenceGenerator(bytesOut);
                seq.addObject(r);
                seq.addObject(s);
                seq.close();
                return bytesOut.toByteArray();
            } catch (IOException e) {
                // Should not happen writing to ByteArrayOutputStream
                throw new UncheckedIOException(e);
            }
        }
    }

    ECDSAPrivateKeySigner(final ECPrivateKey privateKey) {
        this.privateKey = privateKey;
    }

    @Override
    public byte[] sign(final byte[] digest) throws GeneralSecurityException {
        byte[] rawSignature = generateSignature(digest);
        ECSignature signature = decodeSignature(rawSignature);
        signature = preventMalleability(signature);
        return signature.getBytes();
    }

    private byte[] generateSignature(final byte[] digest) throws GeneralSecurityException {
        Signature signer = Signature.getInstance(ALGORITHM_NAME, PROVIDER);
        signer.initSign(privateKey);
        signer.update(digest);
        return signer.sign();
    }

    private ECSignature decodeSignature(final byte[] signature) throws GeneralSecurityException {
        try (ByteArrayInputStream inStream = new ByteArrayInputStream(signature);
             ASN1InputStream asnInputStream = new ASN1InputStream(inStream)) {
            ASN1Primitive asn1 = asnInputStream.readObject();

            if (!(asn1 instanceof ASN1Sequence)) {
                throw new GeneralSecurityException("Invalid signature type: " + asn1.getClass().getSimpleName());
            }

            ASN1Sequence asn1Sequence = (ASN1Sequence) asn1;
            List<ASN1Integer>  signatureParts = StreamSupport.stream(asn1Sequence.spliterator(), false)
                    .map(asn1Encodable -> asn1Encodable.toASN1Primitive())
                    .filter(asn1Primitive -> asn1Primitive instanceof ASN1Integer)
                    .map(asn1Primitive -> (ASN1Integer) asn1Primitive)
                    .collect(Collectors.toList());
            if (signatureParts.size() != 2) {
                throw new GeneralSecurityException("Invalid signature. Expected 2 values but got " + signatureParts.size());
            }

            return new ECSignature(signatureParts.get(0), signatureParts.get(1));
        } catch (IOException e) {
            // Should not happen reading from ByteArrayInputStream
            throw new UncheckedIOException(e);
        }
    }

    private ECSignature preventMalleability(final ECSignature signature) {
        BigInteger s = signature.s.getValue();
        if (s.compareTo(HALF_CURVE_N) > 0) {
            s = CURVE.getN().subtract(s);
            return new ECSignature(signature.r, new ASN1Integer(s));
        }
        return signature;
    }

}
